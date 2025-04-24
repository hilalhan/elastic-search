// Package handlers provides HTTP handlers for the API
package handlers

import (
	"elasticsearch/internal/common"
	"elasticsearch/internal/config"
	"elasticsearch/internal/models"
	"elasticsearch/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productService services.ProductService
	cfg            *config.Config
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(cfg *config.Config, productService services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		cfg:            cfg,
	}
}

// GetProducts handles GET requests to fetch products
// @Summary     Get Products
// @Description Retrieves a list of products with pagination and search keywords
// @Tags        Products
// @Accept      json
// @Produce     json
// @Param       limit   query int false "Limit number of results"
// @Param       offset  query int false "Offset for pagination"
// @Param       keyword query string false "Search keyword"
// @Success 	  200 {object} common.PagedResponse[[]models.Product]
// @Router      /product [get]
func (h *ProductHandler) GetProducts(c fiber.Ctx) error {
	// Parse query parameters
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")
	keyword := c.Query("keyword")

	// Convert string params to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.NewError("Invalid limit parameter", err))
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(common.NewError("Invalid offset parameter", err))
	}

	// Create search parameters
	searchParams := models.ProductSearchParams{
		Limit:   limit,
		Offset:  offset,
		Keyword: keyword,
	}

	// Call service to retrieve products
	result, err := h.productService.GetProducts(c.Context(), searchParams)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(common.NewError("Failed to retrieve products", err))
	}

	// Create pagination info
	pagination := common.PaginationInfo{
		Total:       result.TotalCount,
		Limit:       result.Limit,
		Offset:      result.Offset,
		CurrentPage: result.CurrentPage,
		TotalPages:  result.TotalPages,
	}

	// Return products with pagination info
	response := common.NewPagedSuccess(result.Products, "Products retrieved successfully", pagination)
	return c.JSON(response)
}

// RegisterProductRoutes registers routes for the ProductHandler
func RegisterProductRoutes(app fiber.Router, cfg *config.Config, productService services.ProductService) {
	handler := NewProductHandler(cfg, productService)
	app.Get("/product", handler.GetProducts)
}
