package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"elasticsearch/internal/api"
	"elasticsearch/internal/common"
	"elasticsearch/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gofiber/fiber/v3"
	fiberlog "github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// Application represents the running application and its components
type Application struct {
	config     *config.Config
	fiberApp   *fiber.App
	esClient   *elasticsearch.Client
	shutdownCh chan os.Signal
}

// New creates a new Application instance with the provided configuration
func New(cfg *config.Config) (*Application, error) {
	app := &Application{
		config:     cfg,
		shutdownCh: make(chan os.Signal, 1),
	}

	// Initialize dependencies
	var err error
	if app.esClient, err = initElasticsearch(cfg.Elasticsearch); err != nil {
		return nil, err
	}

	app.fiberApp = initFiber(cfg)

	// Setup routes
	api.RegisterRoute(
		cfg,
		app.fiberApp,
		app.esClient,
	)

	return app, nil
}

// Start begins the server and waits for shutdown signals
func (app *Application) Start() error {
	// Configure graceful shutdown
	signal.Notify(app.shutdownCh, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		addr := app.config.Server.Address
		log.Printf("Starting server on %s", addr)
		if err := app.fiberApp.Listen(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	return app.waitForShutdown()
}

// waitForShutdown blocks until a termination signal is received, then gracefully shuts down the server
func (app *Application) waitForShutdown() error {
	<-app.shutdownCh

	log.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown gracefully with context
	if err := app.fiberApp.ShutdownWithContext(ctx); err != nil {
		return err
	}

	log.Println("Server stopped")
	return nil
}

// initElasticsearch creates and configures a new Elasticsearch client
func initElasticsearch(cfg config.ElasticsearchConfig) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password, // Added password field which was missing
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	// Verify connection
	res, err := es.Info()
	if err != nil {
		return nil, err
	}

	fiberlog.Infof("Connected to Elasticsearch: %v", res.String())
	return es, nil
}

// initFiber creates and configures a new Fiber application
func initFiber(cfg *config.Config) *fiber.App {
	// Create new fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: createErrorHandler(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSec) * time.Second,
	})

	// Apply middleware
	app.Use(
		logger.New(logger.Config{}),
		recover.New(),
	)

	return app
}

// createErrorHandler returns a custom error handler for Fiber
func createErrorHandler() func(c fiber.Ctx, err error) error {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Get specific status code if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Set the status code
		c.Status(code)

		// Return a JSON response with the error details
		return c.JSON(common.NewError(message, err))
	}
}
