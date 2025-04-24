package models

import "time"

// @description Represents a product object
type Product struct {
	ID          uint64    `json:"id"`
	ProductName string    `json:"product_name"`
	DrugGeneric string    `json:"drug_generic"`
	Company     string    `json:"company"`
	Score       float64   `json:"score"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductSearchParams contains parameters for product search
type ProductSearchParams struct {
	Limit   int
	Offset  int
	Keyword string
}

// ProductSearchResult contains products and pagination info
type ProductSearchResult struct {
	Products   []Product
	TotalCount int64
	Limit      int
	Offset     int
}
