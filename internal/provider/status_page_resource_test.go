// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

func TestAccStatusPageResource(t *testing.T) {
	// Skip check.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccStatusPageResourceConfig("test-page", "Test Status Page", "System status page for testing"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("slug"),
						knownvalue.StringExact("test-page"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Test Status Page"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("System status page for testing"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("published"),
						knownvalue.Bool(true),
					),
				},
			},
			// ImportState testing.
			{
				ResourceName:      "uptimekuma_status_page.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: testAccStatusPageResourceConfig("test-page", "Updated Status Page", "Updated description"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("slug"),
						knownvalue.StringExact("test-page"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Updated Status Page"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated description"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase.
		},
	})
}

// Note: This config function definition remains unchanged.
func testAccStatusPageResourceConfig(slug, title, description string) string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  base_url = "%s"
  username = "%s"
  password = "%s"
}

resource "uptimekuma_status_page" "test" {
  slug        = %[4]q
  title       = %[5]q
  description = %[6]q
  published   = true
  theme       = "dark"
  show_tags   = false
}
`,
		os.Getenv("UPTIMEKUMA_BASE_URL"),
		os.Getenv("UPTIMEKUMA_USERNAME"),
		os.Getenv("UPTIMEKUMA_PASSWORD"),
		slug, title, description)
}

func TestAccStatusPageResourceWithGroups(t *testing.T) {
	// Skip check.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with monitor groups.
			{
				Config: testAccStatusPageResourceWithGroupsConfig("test-page-with-groups", "Status Page With Groups"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.with_groups",
						tfjsonpath.New("slug"),
						knownvalue.StringExact("test-page-with-groups"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.with_groups",
						tfjsonpath.New("title"),
						knownvalue.StringExact("Status Page With Groups"),
					),
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.with_groups",
						tfjsonpath.New("public_group_list[0].name"),
						knownvalue.StringExact("Core Services"),
					),
					// Check the name of the second group element [1].
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.with_groups",
						tfjsonpath.New("public_group_list[1].name"),
						knownvalue.StringExact("Secondary Services"),
					),
					// Check weight of first group.
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.with_groups",
						tfjsonpath.New("public_group_list[0].weight"),
						knownvalue.Int64Exact(1),
					),
					// Check weight of second group.
					statecheck.ExpectKnownValue(
						"uptimekuma_status_page.with_groups",
						tfjsonpath.New("public_group_list[1].weight"),
						knownvalue.Int64Exact(2),
					),
				},
			},
			// Delete testing automatically occurs in TestCase.
		},
	})
}

// Note: This config function definition remains unchanged.
func testAccStatusPageResourceWithGroupsConfig(slug, title string) string {
	return fmt.Sprintf(`
provider "uptimekuma" {
  base_url = "%s"
  username = "%s"
  password = "%s"
}

// Define dependent monitors for the group test
resource "uptimekuma_monitor" "http1" {
  name     = "HTTP Monitor 1 for Group Test" // Make names unique for testing
  type     = "http"
  url      = "https://example.com/health"
  interval = 300 // Use longer intervals for Acc tests unless testing interval itself
}

resource "uptimekuma_monitor" "http2" {
  name     = "HTTP Monitor 2 for Group Test" // Make names unique for testing
  type     = "http"
  url      = "https://example.org/status"
  interval = 300
}

resource "uptimekuma_status_page" "with_groups" {
  slug      = %[4]q
  title     = %[5]q
  published = true
  theme     = "dark"

  public_group_list {
    name = "Core Services"
    weight = 1
    monitor_list = [
      uptimekuma_monitor.http1.id // Reference the monitor defined above
    ]
  }

  public_group_list {
    name = "Secondary Services"
    weight = 2
    monitor_list = [
      uptimekuma_monitor.http2.id // Reference the monitor defined above
    ]
  }
}
`,
		os.Getenv("UPTIMEKUMA_BASE_URL"),
		os.Getenv("UPTIMEKUMA_USERNAME"),
		os.Getenv("UPTIMEKUMA_PASSWORD"),
		slug, title)
}
