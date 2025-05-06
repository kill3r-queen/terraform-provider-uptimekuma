# Configure the Uptime Kuma provider
terraform {
  required_providers {
    uptimekuma = {
      source = "kill3r-queen/uptimekuma"
    }
  }
}

provider "uptimekuma" {
  base_url = "http://localhost:3001"  # Your Uptime Kuma instance URL
  username = "admin"                  # Username for authentication
  password = "password"               # Password for authentication
}

# Create HTTP monitors for different service
resource "uptimekuma_monitor" "website" {
  name           = "Company Website"
  type           = "http" 
  url            = "https://example.com"
  interval       = 60
  retry_interval = 30
  max_retries    = 3
}

resource "uptimekuma_monitor" "api" {
  name           = "API Service"
  type           = "http"
  url            = "https://api.example.com/health"
  method         = "GET"
  interval       = 30
  retry_interval = 10
  max_retries    = 3
  ignore_tls     = false
  
  # Optional: Custom headers for API auth
  headers        = "{\"X-API-Key\":\"dummy-key\", \"Accept\":\"application/json\"}"
}

resource "uptimekuma_monitor" "database" {
  name           = "Database Server"
  type           = "ping"
  hostname       = "db.internal.example.com"
  interval       = 60
  retry_interval = 30
  max_retries    = 2
}

# Create a status page with all monitors
resource "uptimekuma_status_page" "main_status" {
  slug        = "status"
  title       = "System Status"
  description = "Current status of Example Company services"
  theme       = "dark"
  published   = true
  
  # Group monitors by service type
  public_group_list = [
    {
      name = "Public Services"
      weight = 1
      monitor_list = [uptimekuma_monitor.website.id]
    },
    {
      name = "API Services"
      weight = 2
      monitor_list = [uptimekuma_monitor.api.id]
    },
    {
      name = "Infrastructure"
      weight = 3
      monitor_list = [uptimekuma_monitor.database.id]
    }
  ]
}