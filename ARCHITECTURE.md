# Terraform Provider Architecture for Uptime Kuma

## Provider Structure

```
terraform-provider-uptimekuma/
├── internal/
│   ├── provider/
│   │   ├── provider.go                     # Provider definition
│   │   ├── resource_monitor.go             # Monitor resource
│   │   ├── resource_status_page.go         # Status page resource
│   │   ├── resource_tag.go                 # Tag resource
│   │   ├── data_source_monitor.go          # Single monitor data source
│   │   ├── data_source_monitors.go         # List monitors data source
│   │   ├── data_source_status_page.go      # Single status page data source
│   │   └── data_source_status_pages.go     # List status pages data source
│   ├── client/
│   │   ├── client.go                       # API client definition
│   │   ├── auth.go                         # Authentication logic
│   │   ├── monitor.go                      # Monitor API operations
│   │   ├── status_page.go                  # Status page API operations
│   │   └── tag.go                          # Tag API operations
│   └── models/
│       ├── monitor.go                      # Monitor data model
│       ├── status_page.go                  # Status page data model
│       └── tag.go                          # Tag data model
```

## Provider Configuration

```hcl
provider "uptimekuma" {
  url             = "https://uptime.example.com"  # Uptime Kuma instance URL
  username        = "admin"                       # Username for authentication
  password        = "password"                    # Password for authentication
  insecure_tls    = false                         # Skip TLS verification (optional)
  request_timeout = 30                            # Request timeout in seconds (optional)
}
```

## Key Resources

### Monitor Resource

```hcl
resource "uptimekuma_monitor" {
  name           = "API Health Check"
  type           = "http"                  # http, ping, port, dns, etc.
  interval       = 60                      # Check interval in seconds
  retry_interval = 60                      # Retry interval
  max_retries    = 3                       # Max retry attempts

  notification_ids = [1, 2]                # Alert notification IDs
  upside_down      = false                 # Invert status logic
  
  # Type-specific configuration (conditional based on type)
  http {
    url                   = "https://api.example.com/health"
    method                = "GET"
    headers               = "Authorization: Bearer token\nContent-Type: application/json"
    body                  = "{\"check\":\"full\"}"
    ignore_tls            = false
    max_redirects         = 10
    accepted_status_codes = [200, 201]
    keyword               = "healthy"      # Keyword to check in response
    
    auth_method        = "basic"           # none, basic, ntlm, mtls
    basic_auth_user    = "user"
    basic_auth_pass    = "pass"
  }
  
  # Other monitor type examples
  
  # ping {
  #   hostname = "example.com"
  # }
  
  # port {
  #   hostname = "db.example.com"
  #   port     = 5432
  # }
  
  # dns {
  #   hostname          = "example.com"
  #   dns_resolve_server = "1.1.1.1"
  #   dns_resolve_type   = "A"
  # }
  
  # Optional tag associations
  tags = [
    {
      tag_id = 1
      value  = "production"
    }
  ]
}
```

### Status Page Resource

```hcl
resource "uptimekuma_status_page" {
  slug        = "main-status"
  title       = "Service Status"
  description = "Current status of our services"
  theme       = "light"                   # light, dark, auto
  published   = true
  
  domain_names  = ["status.example.com"]
  footer_text   = "© 2025 Company"
  show_tags     = true
  custom_css    = ".header { color: blue; }"
  
  # Monitor groups on status page
  public_group {
    name     = "API Services" 
    weight   = 1              # Order of display
    monitors = [1, 2, 3]      # Monitor IDs to include
  }
  
  public_group {
    name     = "Database Services"
    weight   = 2
    monitors = [4, 5]
  }
}
```

### Tag Resource

```hcl
resource "uptimekuma_tag" {
  name  = "production"
  color = "#00FF00"
}
```

## Data Sources

### Monitor Data Source

```hcl
data "uptimekuma_monitor" {
  monitor_id = 1
}
```

### Status Page Data Source

```hcl
data "uptimekuma_status_page" {
  slug = "main-status"
}
```

## Authentication Implementation

The Uptime Kuma API uses OAuth2 Password Grant flow for authentication. The provider will need to:

1. Obtain an access token using username/password credentials
2. Use the token for all subsequent API requests
3. Handle token refresh when expired

```go
// Example authentication flow based on API requirements
func (c *Client) Authenticate(ctx context.Context) error {
    // Use the /login/access-token endpoint with form-encoded credentials
    data := url.Values{}
    data.Set("username", c.username)
    data.Set("password", c.password)
    
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        fmt.Sprintf("%s/api/v1/login/access-token", c.baseURL),
        strings.NewReader(data.Encode()),
    )
    if err != nil {
        return fmt.Errorf("error creating authentication request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    var tokenResponse struct {
        AccessToken string `json:"access_token"`
        TokenType   string `json:"token_type"`
    }
    
    // Send request and parse response
    if err := c.sendRequest(req, &tokenResponse); err != nil {
        return fmt.Errorf("error authenticating with Uptime Kuma: %w", err)
    }
    
    // Store the token for use in subsequent requests
    c.token = tokenResponse.AccessToken
    return nil
}

// AddAuthHeaders adds the authorization header to all API requests
func (c *Client) AddAuthHeaders(req *http.Request) {
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }
}
```

Example of authentication via curl:
```bash
# Get auth token
TOKEN=$(curl -X -L 'POST' -H 'Content-Type: application/x-www-form-urlencoded' \
  --data 'username=admin&password=admin' \
  http://127.0.0.1:8000/login/access-token/ | jq -r ".access_token")

# Use token to authenticate API requests
curl -L -H 'Accept: application/json' -H "Authorization: Bearer ${TOKEN}" \
  http://127.0.0.1:8000/monitors/
```

## Implementation Strategy

1. Start with provider framework and authentication
   - Implement OAuth2 token-based authentication as shown above
   - Handle token refresh and error handling

2. Implement the monitor resource (core functionality)
   - Create all monitor type schemas
   - Build CRUD operations for monitors
   - Implement monitor pause/resume functionality

3. Add status page resource
   - Status page creation and management
   - Support for multiple public groups
   - Integration with monitors

4. Implement tag resources
   - Tag management for monitors
   - Color validation

5. Add data sources for read-only operations
   - Read monitors and status pages
   - List all resources

6. Create comprehensive acceptance tests
   - Test all CRUD operations
   - Test resource relationships
   - Test error handling

## Design Considerations

- Use Terraform Plugin Framework for modern provider development
- Handle different monitor types with conditional schema validation
- Implement proper OAuth2 token management with token refresh
- Use plan modifiers for sensitive fields like passwords
- Support importing existing resources for all resource types
- Add data validators for all resource attributes
- Implement custom diffs for complex nested structures
- Consider retry logic for API rate limits or temporary failures