# Elasticsearch Go Monolith Application

This project implements a Go application with Elasticsearch integration using a monolithic architecture. The structure is designed to be easily converted to microservices in the future.

## Project Structure

```
elastic-search/
├── cmd/
│   └── server/
│       └── main.go             # Application entry point
├── docs/                       # API documentation
│   ├── docs.go
│   ├── swagger.json            # Swagger JSON specification
│   └── swagger.yaml            # Swagger YAML specification
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── common/
│   │   └── response.go         # Common response utilities
│   ├── api/
│   │   ├── handlers/           # HTTP request handlers
│   │   │   ├── health.go       # Health check handler
│   │   │   └── product.go      # Product handlers
│   │   └── routes.go           # API route definitions
│   ├── app/
│   │   ├── application.go      # Application setup
│   │   └── importer.go         # Data import functionality
│   ├── models/
│   │   └── product.go          # Product data structures
│   ├── storage/
│   │   └── elasticsearch/
│   │       ├── client.go       # Elasticsearch connection management
│   │       ├── importer.go     # Data import implementation
│   │       └── repository.go   # Data access layer
│   └── services/
│       └── product.go          # Product business logic
├── pkg/
│   └── shared/                 # Reusable utilities
├── Dockerfile                  # Container definition
├── docker-compose.yml          # Multi-container setup
├── .env.example                # Example environment variables
├── go.mod                      # Module dependencies
└── go.sum                      # Dependency checksums
```

## Directory Explanations

### `/cmd`

Contains the executable entry points for the application.

- **`/cmd/server/main.go`**: The main application function that:
  - Loads configuration
  - Initializes components (Elasticsearch client, repository, service, handlers)
  - Sets up HTTP server
  - Implements graceful shutdown

### `/docs`

Contains API documentation generated with Swagger.

- **`docs.go`**: Go source file for Swagger documentation
- **`swagger.json`**: Swagger documentation in JSON format
- **`swagger.yaml`**: Swagger documentation in YAML format

### `/internal`

Contains code that's specific to this application and not intended to be imported by other projects.

#### `/internal/config`

- **`config.go`**: Defines and loads application configuration from environment variables
- **Scope**: Manages all configuration settings needed by the application

#### `/internal/common`

- **`response.go`**: Common HTTP response utilities
- **Scope**: Standardizes API responses across the application

#### `/internal/api`

Contains all HTTP API-related code.

- **`/handlers`**: HTTP request handlers
  - **`health.go`**: Implements health check endpoints
  - **`product.go`**: Implements product-related endpoints
  - **Scope**: Converts HTTP requests to service calls and formats responses
- **`routes.go`**: API endpoint definitions
  - **Scope**: Maps URLs to handler functions and applies middleware

#### `/internal/models`

- **`product.go`**: Product data structures
- **Scope**: Defines the domain objects used throughout the application

#### `/internal/storage`

Handles data persistence concerns.

- **`/elasticsearch`**: Elasticsearch-specific implementation
  - **`client.go`**: Manages connections to Elasticsearch
    - **Scope**: Connection setup, health checks, cluster operations
  - **`importer.go`**: Implements data import functionality
    - **Scope**: Handles importing data from external sources
  - **`repository.go`**: Data access patterns
    - **Scope**: CRUD operations for specific indices

#### `/internal/services`

- **`product.go`**: Product business logic implementation
- **Scope**: Orchestrates operations, applies business rules, uses repositories

### `/pkg`

Contains code that could be reused by external applications.

- **`/shared`**: Reusable utilities and helpers
  - **Scope**: Generic functions not specific to this application

## How Components Work Together

1. The **main function** loads configuration from environment variables
2. Initializes the Elasticsearch **client**
3. Client creates a **repository**
4. Repository is used by the **service** layer
5. Service layer is used by API **handlers**
6. Handlers are registered with **routes**
7. HTTP server starts with the configured router

## Benefits of This Architecture

- **Separation of Concerns**: Each component has a single responsibility
- **Testability**: Components can be tested in isolation
- **Maintainability**: Changes in one layer don't affect others
- **Extensibility**: Easy to add new features
- **Microservice-Ready**: Components can be extracted into separate services

## Microservice Migration Path

When ready to migrate to microservices:

1. Extract services into separate repositories
2. Add API gateway to route requests
3. Implement service discovery
4. Add message broker for asynchronous communication
5. Implement distributed tracing

## Getting Started

1. Clone the repository
2. Create `.env` file from `.env.example` (see example below)
3. Run Elasticsearch
4. Build and run the application

### Environment Configuration (.env)

Copy the `.env.example` file to create your own `.env` file:

```
ENVIRONMENT=development

# Application
SERVER_ADDRESS=:8080
SERVER_READ_TIMEOUT_SEC=15
SERVER_WRITE_TIMEOUT_SEC=15
SERVER_IDLE_TIMEOUT_SEC=60

# Elasticsearch
# separate multiple addresses with commas (e.g. http://localhost:9200,http://localhost:9201)
ELASTICSEARCH_ADDRESSES=http://localhost:9200
ELASTICSEARCH_USERNAME=
ELASTICSEARCH_PASSWORD=
ELASTICSEARCH_INDEX=items
ELASTICSEARCH_TIMEOUT_SEC=5
```

### Running with Docker Compose

```bash
docker-compose up
```

### Generating Swagger Documentation

This project uses [swaggo](https://github.com/swaggo/swag) to generate Swagger documentation. To generate or update the Swagger documentation, run:

```bash
# Install swaggo if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swag init --generalInfo cmd/server/main.go --dir ./
```

This command will scan your API annotations in the code and generate updated documentation files in the `/docs` directory.

### Import Data

```bash
docker compose run app --import-excel="https://docs.google.com/spreadsheets/d/191toBNpYauM-gA36MsVfgUMCg4LpWKqShvXf6K7C8MY/edit?usp=sharing"
```
