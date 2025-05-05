// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0.

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMonitorResource(t *testing.T) {
	initialDescription := "Test HTTP Monitor Description"
	updatedDescription := "Updated Test HTTP Monitor Description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorResourceConfig("HTTP Monitor", "http", "https://example.com", initialDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("HTTP Monitor"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("http"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact("https://example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(initialDescription),
					),
				},
			},
			{
				ResourceName:      "uptimekuma_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,

				ImportStateVerifyIgnore: []string{"basic_auth_pass"},
			},
			{
				Config: testAccMonitorResourceConfig("Updated HTTP Monitor", "http", "https://updated-example.com", updatedDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("Updated HTTP Monitor"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("url"),
						knownvalue.StringExact("https://updated-example.com"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(updatedDescription),
					),
				},
			},
		},
	})
}

func testAccMonitorResourceConfig(name, monitorType, url, description string) string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  base_url = "%s"
  username = "%s"
  password = "%s"
}

resource "uptimekuma_monitor" "test" {
  name            = %[4]q // Index 4 = name
  type            = %[5]q // Index 5 = monitorType
  url             = %[6]q // Index 6 = url
  description     = %[7]q // Index 7 = description (Keep this line)
  interval        = 60
  max_retries     = 3
  retry_interval  = 30
}
`,
		os.Getenv("UPTIMEKUMA_BASE_URL"),
		os.Getenv("UPTIMEKUMA_USERNAME"),
		os.Getenv("UPTIMEKUMA_PASSWORD"),
		name, monitorType, url, description)
}

func TestAccPingMonitorResource(t *testing.T) {
	// CHANGE HERE: Define a description for the ping test.
	pingDescription := "Test Ping Description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// CHANGE HERE: Pass 'pingDescription' as the third argument (line 112 in original error).
				Config: testAccPingMonitorResourceConfig("Ping Monitor", "example.com", pingDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.ping_test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("Ping Monitor"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.ping_test",
						tfjsonpath.New("type"),
						knownvalue.StringExact("ping"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.ping_test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact("example.com"),
					),
					// CHANGE HERE: Add state check for description.
					statecheck.ExpectKnownValue(
						"uptimekuma_monitor.ping_test",
						tfjsonpath.New("description"),
						knownvalue.StringExact(pingDescription),
					),
				},
			},
		},
	})
}

// Keep the testAccPingMonitorResourceConfig function definition as you already had it.
func testAccPingMonitorResourceConfig(name, hostname, description string) string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  base_url = "%s"
  username = "%s"
  password = "%s"
}

resource "uptimekuma_monitor" "ping_test" {
  name            = %[4]q
  type            = "ping"
  hostname        = %[5]q
  description     = %[6]q
  interval        = 60
  max_retries     = 3
}
`,
		os.Getenv("UPTIMEKUMA_BASE_URL"),
		os.Getenv("UPTIMEKUMA_USERNAME"),
		os.Getenv("UPTIMEKUMA_PASSWORD"),
		name, hostname, description)
}
