// Package services provides business logic implementation
package services

import (
	"context"
	"elasticsearch/internal/models"
	"elasticsearch/internal/storage/elasticsearch"
	"math"
)

type ProductSearchResult struct {
	Products    []models.Product
	TotalCount  int64
	Limit       int
	Offset      int
	CurrentPage int
	TotalPages  int
}

type ProductService interface {
	GetProducts(ctx context.Context, params models.ProductSearchParams) (ProductSearchResult, error)
}

type ProductServiceImpl struct {
	productRepo elasticsearch.ProductRepository
}

func NewProductService(productRepo elasticsearch.ProductRepository) *ProductServiceImpl {
	return &ProductServiceImpl{
		productRepo: productRepo,
	}
}

func (s *ProductServiceImpl) GetProducts(ctx context.Context, params models.ProductSearchParams) (ProductSearchResult, error) {
	// Call repository to get products
	result, err := s.productRepo.FindProducts(ctx, params)
	if err != nil {
		return ProductSearchResult{}, err
	}

	// Calculate page info
	currentPage := 1
	if params.Limit > 0 {
		currentPage = (params.Offset / params.Limit) + 1
	}

	totalPages := 1
	if params.Limit > 0 && result.TotalCount > 0 {
		totalPages = int(math.Ceil(float64(result.TotalCount) / float64(params.Limit)))
	}

	// Return products with pagination info
	return ProductSearchResult{
		Products:    result.Products,
		TotalCount:  result.TotalCount,
		Limit:       params.Limit,
		Offset:      params.Offset,
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}, nil
}
