package api

import (
	"elasticsearch/internal/api/handlers"
	"elasticsearch/internal/config"
	"elasticsearch/internal/services"
	"log"
	"os"

	storageEs "elasticsearch/internal/storage/elasticsearch"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v3"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func RegisterRoute(cfg *config.Config, app *fiber.App, es *elasticsearch.Client) {
	// Create repositories
	productRepo := storageEs.NewElasticsearchProductRepository(es, "products")

	// Create services
	productService := services.NewProductService(productRepo)

	// Create handlers
	app.Get("/docs/swagger.json", func(c fiber.Ctx) error {
		file, err := os.ReadFile("./docs/swagger.json")
		if err != nil {
			log.Println("Failed to read swagger.json:", err)
			return c.Status(500).SendString("Failed to load swagger.json")
		}
		return c.Send(file)
	})

	app.Get("/swagger/*", func(c fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/docs/swagger.json")))(c.Context())
		return nil
	})

	app.Get("/health", handlers.Health)
	handlers.RegisterProductRoutes(app, cfg, productService)
}
