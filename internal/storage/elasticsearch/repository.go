package elasticsearch

import (
	"bytes"
	"context"
	"elasticsearch/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	FindProducts(ctx context.Context, params models.ProductSearchParams) (models.ProductSearchResult, error)
}

// ElasticsearchProductRepository implements ProductRepository using Elasticsearch
type ElasticsearchProductRepository struct {
	es        *elasticsearch.Client
	indexName string
}

// NewElasticsearchProductRepository creates a new ElasticsearchProductRepository
func NewElasticsearchProductRepository(es *elasticsearch.Client, indexName string) *ElasticsearchProductRepository {
	return &ElasticsearchProductRepository{
		es:        es,
		indexName: indexName,
	}
}

// FindProducts retrieves products from Elasticsearch based on search parameters
func (r *ElasticsearchProductRepository) FindProducts(ctx context.Context, params models.ProductSearchParams) (models.ProductSearchResult, error) {
	// Build the elasticsearch query
	query := r.buildProductQuery(params)

	// Encode query to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Printf("Error encoding query: %s", err)
		return models.ProductSearchResult{}, fmt.Errorf("failed to encode query: %w", err)
	}

	// Perform the search request
	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex(r.indexName),
		r.es.Search.WithBody(&buf),
		r.es.Search.WithTrackTotalHits(true),
		r.es.Search.WithPretty(),
	)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return models.ProductSearchResult{}, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	// Check for Elasticsearch errors
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return models.ProductSearchResult{}, fmt.Errorf("error parsing elasticsearch error response: %w", err)
		}

		errorMsg := fmt.Sprintf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
		log.Print(errorMsg)
		return models.ProductSearchResult{}, fmt.Errorf(errorMsg)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		log.Printf("Error parsing response body: %s", err)
		return models.ProductSearchResult{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract products from response
	products, err := r.extractProductsFromResponse(response)
	if err != nil {
		return models.ProductSearchResult{}, fmt.Errorf("failed to extract products from response: %w", err)
	}

	// Extract total count
	totalCount := r.extractTotalCount(response)

	// Create and return search result with pagination info
	result := models.ProductSearchResult{
		Products:   products,
		TotalCount: totalCount,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}

	return result, nil
}

// extractTotalCount extracts the total hit count from Elasticsearch response
func (r *ElasticsearchProductRepository) extractTotalCount(response map[string]interface{}) int64 {
	hits, ok := response["hits"].(map[string]interface{})
	if !ok {
		return 0
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return 0
	}

	value, ok := total["value"].(float64)
	if !ok {
		return 0
	}

	return int64(value)
}

// buildProductQuery constructs the Elasticsearch query based on search parameters
func (r *ElasticsearchProductRepository) buildProductQuery(params models.ProductSearchParams) map[string]interface{} {
	query := map[string]interface{}{
		"from": params.Offset,
		"size": params.Limit,
	}

	// Add search conditions if keyword is provided
	if params.Keyword != "" {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []map[string]interface{}{
						{
							"bool": map[string]interface{}{
								"should": []map[string]interface{}{
									{
										"match": map[string]interface{}{
											"product_name": map[string]interface{}{
												"query":     params.Keyword,
												"operator":  "and",
												"fuzziness": "AUTO",
											},
										},
									},
									{
										"match": map[string]interface{}{
											"drug_generic": map[string]interface{}{
												"query":     params.Keyword,
												"operator":  "and",
												"fuzziness": "AUTO",
											},
										},
									},
									{
										"match": map[string]interface{}{
											"company": map[string]interface{}{
												"query":     params.Keyword,
												"operator":  "and",
												"fuzziness": "AUTO",
											},
										},
									},
								},
							},
						},
						{
							"bool": map[string]interface{}{
								"should": []map[string]interface{}{
									{
										"wildcard": map[string]interface{}{
											"product_name": map[string]interface{}{
												"value": "*" + params.Keyword + "*",
											},
										},
									},
									{
										"wildcard": map[string]interface{}{
											"drug_generic": map[string]interface{}{
												"value": "*" + params.Keyword + "*",
											},
										},
									},
									{
										"wildcard": map[string]interface{}{
											"company": map[string]interface{}{
												"value": "*" + params.Keyword + "*",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"sort": []map[string]interface{}{
				{"_score": map[string]interface{}{"order": "desc"}},
				{"product_name.keyword": map[string]interface{}{"order": "asc"}},
			},
			"from": params.Offset,
			"size": params.Limit,
		}
	}

	return query
}

func (r *ElasticsearchProductRepository) extractProductsFromResponse(response map[string]interface{}) ([]models.Product, error) {
	products := []models.Product{}

	hits, ok := response["hits"].(map[string]interface{})["hits"]
	if !ok {
		return products, nil // Return empty slice if no hits field
	}

	hitsArray, ok := hits.([]interface{})
	if !ok {
		return products, nil // Return empty slice if hits is not an array
	}

	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		docId := hitMap["_id"]
		score := hitMap["_score"]
		source := hitMap["_source"]

		jsonData, err := json.Marshal(source)
		if err != nil {
			log.Printf("Error marshaling hit source: %s", err)
			continue
		}

		var product models.Product
		if err := json.Unmarshal(jsonData, &product); err != nil {
			log.Printf("Error unmarshaling product: %s", err)
			continue
		}

		// Set ID and score
		if idStr, ok := docId.(string); ok {
			if id, err := strconv.Atoi(idStr); err == nil {
				product.ID = uint64(id)
			}
		}

		if scoreFloat, ok := score.(float64); ok {
			product.Score = scoreFloat
		}

		products = append(products, product)
	}

	return products, nil
}
