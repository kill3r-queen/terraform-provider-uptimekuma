# Terraform Provider for Uptime Kuma

This Terraform provider allows you to manage [Uptime Kuma](https://github.com/louislam/uptime-kuma) resources through Terraform. Uptime Kuma is a self-hosted monitoring tool similar to Uptime Robot.

## Features

- **Monitors**: Create and manage HTTP, Ping, Port, DNS, and other monitor types
- **Status Pages**: Create and manage status pages with custom domains
- **Tags**: Organize monitors with tags

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23
- Access to an Uptime Kuma instance (self-hosted or hosted)

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

Configure the provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    uptimekuma = {
      source = "yourusername/uptimekuma"
    }
  }
}

provider "uptimekuma" {
  url      = "https://uptime.example.com"  # Your Uptime Kuma instance URL
  username = "admin"                       # Username for authentication
  password = "password"                    # Password for authentication
}

# Create an HTTP monitor
resource "uptimekuma_monitor" "api" {
  name           = "API Health Check"
  type           = "http"
  interval       = 60
  retry_interval = 60
  max_retries    = 3
  
  http {
    url                   = "https://api.example.com/health"
    method                = "GET"
    accepted_status_codes = [200, 201]
  }
}

# Create a status page
resource "uptimekuma_status_page" "main" {
  slug        = "main-status"
  title       = "Service Status"
  description = "Current status of our services"
  theme       = "light"
  published   = true
  
  domain_names = ["status.example.com"]
  
  public_group {
    name     = "API Services"
    weight   = 1
    monitors = [uptimekuma_monitor.api.id]
  }
}
```

See the [docs](./docs/) directory for more detailed documentation on available resources and their options.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

### Setting Up Local Environment

1. Clone the repository
2. Install dependencies: `go mod download`

### Running Tests

This provider includes both unit tests and acceptance tests.

**Running Unit Tests**

Unit tests run against mock API endpoints and don't require a real Uptime Kuma instance:

```shell
# Run all unit tests
go test -v ./...

# Run specific test
go test -v ./internal/client -run TestAuthentication

# Run tests with coverage
go test -v -cover ./...
```

**Running Acceptance Tests**

Acceptance tests create real resources on your Uptime Kuma instance. These require an actual Uptime Kuma instance and valid credentials:

```shell
# Set required environment variables
export UPTIMEKUMA_URL="http://localhost:3001"
export UPTIMEKUMA_USERNAME="admin"
export UPTIMEKUMA_PASSWORD="mypassword"

# Run acceptance tests
make testacc
```

*Note:* Acceptance tests create and destroy real resources. Use with caution on production instances.

### Generate Documentation

To generate or update documentation, run:

```shell
make generate
```

## Architecture

The provider is built using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and follows a clean architecture pattern:

1. **Client Layer** (`internal/client/`): API client handling authentication and API calls
2. **Provider Layer** (`internal/provider/`): Terraform resource and data source implementations
3. **Models** (`internal/models/`): Data structures shared between layers

For more details, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## API Documentation

For details on the Uptime Kuma API that this provider uses, see [API.md](./API.md).

## License

[MPL 2.0](LICENSE)