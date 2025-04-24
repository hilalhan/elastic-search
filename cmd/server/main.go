package main

import (
	"flag"
	"os"

	"elasticsearch/internal/app"
	"elasticsearch/internal/config"

	fiberlog "github.com/gofiber/fiber/v3/log"
)

// @title Elastic Search Skill-Test
// @version 1.0
// @description This is a swagger for the service
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func main() {
	// Parse command-line flags
	flags := parseFlags()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fiberlog.Fatalf("Failed to load configuration: %v", err)
	}

	// Handle import mode if specified
	if flags.importPath != "" {
		if err := executeImport(cfg, flags.importPath); err != nil {
			fiberlog.Fatalf("❌ Import failed: %v", err)
		}
		return
	}

	// Run the application in server mode
	if err := startServer(cfg); err != nil {
		fiberlog.Fatalf("Application error: %v", err)
		os.Exit(1)
	}
}

// executeImport handles importing data from Excel
func executeImport(cfg *config.Config, path string) error {
	fiberlog.Infof("Starting import from: %s", path)
	return app.ImportExcel(cfg, path)
}

// startServer initializes and starts the application server
func startServer(cfg *config.Config) error {
	// Initialize the application
	application, err := app.New(cfg)
	if err != nil {
		return err
	}

	// Start the server (this is a blocking call that waits for shutdown)
	fiberlog.Info("Starting application server...")
	return application.Start()
}

// CommandFlags holds all command-line flags
type CommandFlags struct {
	importPath string
}

// parseFlags parses command-line arguments and returns structured flags
func parseFlags() CommandFlags {
	var flags CommandFlags

	flag.StringVar(&flags.importPath, "import-excel", "", "Path to Excel file to import")
	flag.Parse()

	return flags
}
