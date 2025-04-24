package elasticsearch

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ESClient struct {
	*elasticsearch.Client
}

type Config struct {
	Addresses []string
	Username  string
	Password  string
	APIKey    string
	Timeout   time.Duration
}

func NewClient(cfg Config) (*ESClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		APIKey:    cfg.APIKey,
	}

	// Set timeout if provided
	if cfg.Timeout > 0 {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.ResponseHeaderTimeout = cfg.Timeout
		esCfg.Transport = transport
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	// Test the connection
	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	log.Println("Connected to Elasticsearch!")

	return &ESClient{Client: client}, nil
}

// Health performs a cluster health check
func (c *ESClient) Health(ctx context.Context) (string, error) {
	res, err := c.Cluster.Health(
		c.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return "", err
	}

	return health["status"].(string), nil
}
