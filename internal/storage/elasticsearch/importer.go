package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"elasticsearch/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	fiberlog "github.com/gofiber/fiber/v3/log"
)

// ImportFromExcel imports data from an Excel file or Google Sheets URL
func ImportFromExcel(esClient *elasticsearch.Client, indexName string, filePath string) error {
	// Check if the path is a Google Sheets URL
	if strings.Contains(filePath, "docs.google.com/spreadsheets") {
		return importFromGoogleSheets(esClient, indexName, filePath)
	}

	// Handle local file import (implementation would be similar but using excelize)
	return fmt.Errorf("local file import not implemented")
}

// importFromGoogleSheets imports data from a Google Sheets URL
func importFromGoogleSheets(esClient *elasticsearch.Client, indexName string, sheetsURL string) error {
	// Extract the spreadsheet ID from the URL
	spreadsheetID, err := extractSpreadsheetID(sheetsURL)
	if err != nil {
		return err
	}

	// Download the CSV data
	csvData, err := downloadGoogleSheetCSV(spreadsheetID)
	if err != nil {
		return err
	}

	// Parse CSV data
	lines := strings.Split(csvData, "\n")
	if len(lines) < 2 {
		return fmt.Errorf("spreadsheet contains no data")
	}

	// Process header and validate columns
	columnMap, err := validateCSVHeaders(lines[0])
	if err != nil {
		return err
	}

	// Create index if it doesn't exist
	err = createIndexIfNotExists(esClient, indexName)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Process data lines and create products
	products := processCSVDataLines(lines, columnMap)

	// Import products in batches using bulk API
	return importProductsBulk(esClient, indexName, products)
}

// downloadGoogleSheetCSV downloads CSV data from Google Sheets
func downloadGoogleSheetCSV(spreadsheetID string) (string, error) {
	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", spreadsheetID)
	fiberlog.Infof("Downloading spreadsheet data from: %s", exportURL)

	resp, err := http.Get(exportURL)
	if err != nil {
		return "", fmt.Errorf("failed to download spreadsheet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download spreadsheet, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// validateCSVHeaders validates that required columns exist in the CSV
func validateCSVHeaders(headerLine string) (map[string]int, error) {
	headerFields := parseCSVLine(headerLine)
	columnMap := make(map[string]int)
	requiredColumns := []string{"id", "product_name", "drug_generic", "company"}

	// Map column names to indices
	for i, header := range headerFields {
		columnMap[strings.ToLower(header)] = i
	}

	// Verify all required columns exist
	for _, col := range requiredColumns {
		if _, exists := columnMap[col]; !exists {
			return nil, fmt.Errorf("required column '%s' not found in spreadsheet", col)
		}
	}

	return columnMap, nil
}

// processCSVDataLines processes CSV data lines into Product objects
func processCSVDataLines(lines []string, columnMap map[string]int) []models.Product {
	var products []models.Product
	now := time.Now()
	requiredColumns := []string{"id", "product_name", "drug_generic", "company"}

	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			continue // Skip empty lines
		}

		fields := parseCSVLine(line)
		if len(fields) < len(requiredColumns) {
			fiberlog.Warnf("Row %d has fewer fields than expected, skipping", i+1)
			continue
		}

		// Parse ID to uint64
		idStr := fields[columnMap["id"]]
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			fiberlog.Warnf("Invalid ID at row %d: %v, skipping", i+1, err)
			continue
		}

		product := models.Product{
			ID:          id,
			ProductName: fields[columnMap["product_name"]],
			DrugGeneric: fields[columnMap["drug_generic"]],
			Company:     fields[columnMap["company"]],
			Score:       0.0, // Default score
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		products = append(products, product)
	}

	return products
}

// parseCSVLine properly handles CSV lines, considering quoted values that might contain commas
func parseCSVLine(line string) []string {
	var result []string
	var currentField strings.Builder
	inQuotes := false

	for _, char := range line {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				currentField.WriteRune(char)
			} else {
				result = append(result, currentField.String())
				currentField.Reset()
			}
		default:
			currentField.WriteRune(char)
		}
	}

	// Add the last field
	result = append(result, currentField.String())

	// Clean up fields (remove surrounding quotes and whitespace)
	for i := range result {
		result[i] = strings.Trim(result[i], "\" \t\r\n")
	}

	return result
}

// importProductsBulk imports products using the Elasticsearch bulk API
func importProductsBulk(esClient *elasticsearch.Client, indexName string, products []models.Product) error {
	if len(products) == 0 {
		fiberlog.Info("No products to import")
		return nil
	}

	fiberlog.Infof("Starting bulk import of %d products", len(products))

	// Create a bulk request
	var bulkBody strings.Builder
	batchSize := 100

	for i, product := range products {
		// Add bulk action - using string ID for Elasticsearch
		actionLine := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%d"}}`, indexName, product.ID)
		bulkBody.WriteString(actionLine)
		bulkBody.WriteString("\n")

		// Add document data
		productJSON, err := json.Marshal(product)
		if err != nil {
			fiberlog.Warnf("Failed to marshal product %d: %v", product.ID, err)
			continue
		}

		bulkBody.Write(productJSON)
		bulkBody.WriteString("\n")

		// Process in batches
		if (i+1)%batchSize == 0 || i == len(products)-1 {
			// Send the batch
			req := esapi.BulkRequest{
				Body: strings.NewReader(bulkBody.String()),
			}

			res, err := req.Do(context.Background(), esClient)
			if err != nil {
				fiberlog.Errorf("Bulk request failed: %v", err)
				continue
			}

			if res.IsError() {
				responseBody, _ := io.ReadAll(res.Body)
				fiberlog.Errorf("Bulk request returned error: %s", string(responseBody))
			} else {
				fiberlog.Infof("Successfully processed batch of %d products", min(batchSize, len(products)-i+batchSize-1))
			}

			res.Body.Close()
			bulkBody.Reset()
		}
	}

	fiberlog.Info("✅ Bulk import completed")
	return nil
}

// createIndexIfNotExists creates the Elasticsearch index if it doesn't already exist
func createIndexIfNotExists(esClient *elasticsearch.Client, indexName string) error {
	// Check if index exists
	res, err := esClient.Indices.Exists([]string{indexName})
	if err != nil {
		return err
	}

	// If index exists, return
	if res.StatusCode == 200 {
		return nil
	}

	// Create index with mapping for our Product struct
	mapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "long"},
				"product_name": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"drug_generic": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"company": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"score": {"type": "float"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"}
			}
		}
	}`

	res, err = esClient.Indices.Create(
		indexName,
		esClient.Indices.Create.WithBody(strings.NewReader(mapping)),
	)

	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	return nil
}

// extractSpreadsheetID extracts the Google Sheets ID from a URL
func extractSpreadsheetID(url string) (string, error) {
	// Pattern for Google Sheets URLs
	r := regexp.MustCompile(`/d/([a-zA-Z0-9-_]+)`)
	matches := r.FindStringSubmatch(url)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract spreadsheet ID from URL")
	}

	return matches[1], nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
