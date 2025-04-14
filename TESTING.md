# Testing the Uptime Kuma Terraform Provider

This document provides guidance on how to test the Uptime Kuma Terraform Provider.

## Test Types

The provider includes two types of tests:

1. **Unit Tests**: Tests individual functions and methods without external dependencies
2. **Acceptance Tests**: Tests that create real resources against a real Uptime Kuma instance

## Running Tests

### Unit Tests

Unit tests can be run with the standard Go test command:

```bash
go test -v ./internal/client
go test -v ./internal/provider
```

### Acceptance Tests

Acceptance tests create real resources and require a running Uptime Kuma instance. These tests are controlled by the `TF_ACC` environment variable.

```bash
# Run all acceptance tests
TF_ACC=1 go test -v ./internal/provider

# Run a specific test
TF_ACC=1 go test -v -run=TestAccMonitorResource ./internal/provider
```

For acceptance tests, you need to set the following environment variables:

```bash
export TF_ACC=1
export UPTIMEKUMA_BASE_URL="http://localhost:3001"
export UPTIMEKUMA_USERNAME="admin"
export UPTIMEKUMA_PASSWORD="password"
```

You can also run the tests with the Makefile command:

```bash
make testacc
```

## Writing Tests

### Unit Tests

Follow standard Go testing practices for unit tests. For testing client functionality, look at the example test files in the `internal/client` directory.

### Acceptance Tests

Acceptance tests use the Terraform testing framework. Here's how to structure an acceptance test:

```go
func TestAccResourceName(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccResourceNameConfig("attribute-value"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "uptimekuma_resource_name.test",
                        tfjsonpath.New("attribute_name"),
                        knownvalue.StringExact("attribute-value"),
                    ),
                },
            },
            // ImportState testing
            {
                ResourceName:      "uptimekuma_resource_name.test",
                ImportState:       true,
                ImportStateVerify: true,
                // If some fields cannot be imported, add them here
                ImportStateVerifyIgnore: []string{"sensitive_attribute"},
            },
            // Update and Read testing
            {
                Config: testAccResourceNameConfig("updated-value"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "uptimekuma_resource_name.test",
                        tfjsonpath.New("attribute_name"),
                        knownvalue.StringExact("updated-value"),
                    ),
                },
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}

// Helper function to generate Terraform configuration
func testAccResourceNameConfig(attributeValue string) string {
    return fmt.Sprintf(`
provider "uptimekuma" {
  base_url = "http://localhost:3001"
  username = "admin"
  password = "password"
}

resource "uptimekuma_resource_name" "test" {
  attribute_name = %[1]q
}
`, attributeValue)
}
```

## Testing Different Scenarios

When testing resources, consider different scenarios:

1. **Basic creation and reading** - Verify that a resource can be created and read correctly
2. **Import testing** - Verify that existing resources can be imported
3. **Update testing** - Verify that resources can be updated
4. **Attribute validation** - Test how the provider handles invalid attributes
5. **Error handling** - Test how the provider handles API errors

## Mocking

For complex tests, consider using mocks to simulate the Uptime Kuma API. This can make tests more reliable and faster to run, especially when you want to test edge cases or error conditions.

Example of a simple mock server:

```go
func setupMockServer() *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case r.Method == "GET" && r.URL.Path == "/monitors":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`[{"id": 1, "name": "Example", "type": "http"}]`))
        // Add more case handlers as needed
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
}
```

## State Checking

The test framework provides powerful state checking capabilities to verify the actual Terraform state after each operation. Use these to ensure resources are created and updated correctly:

```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "uptimekuma_monitor.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("Test Monitor"),
    ),
    statecheck.ExpectKnownValue(
        "uptimekuma_monitor.test",
        tfjsonpath.New("type"),
        knownvalue.StringExact("http"),
    ),
},
```

For more complex state checking, use the full range of matchers provided by the testing framework.