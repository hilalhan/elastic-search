package app

import (
	"time"

	"elasticsearch/internal/config"
	"elasticsearch/internal/storage/elasticsearch"

	fiberlog "github.com/gofiber/fiber/v3/log"
)

// ImportExcel handles importing data from an Excel file into Elasticsearch
func ImportExcel(cfg *config.Config, importPath string) error {
	// Create temporary client for import
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Elasticsearch.Addresses,
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
		Timeout:   time.Duration(cfg.Elasticsearch.TimeoutSec) * time.Second,
	})
	if err != nil {
		return err
	}

	fiberlog.Info("📥 Importing spreadsheet from", importPath, "with index:", cfg.Elasticsearch.Index)
	if err := elasticsearch.ImportFromExcel(esClient.Client, cfg.Elasticsearch.Index, importPath); err != nil {
		return err
	}

	fiberlog.Info("✅ Import complete")
	return nil
}
