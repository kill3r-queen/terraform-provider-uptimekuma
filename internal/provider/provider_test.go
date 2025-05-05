// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	// "github.com/hashicorp/terraform-plugin-testing/echoprovider" // Keep if needed elsewhere, remove if only used by the deleted variable.
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"uptimekuma": providerserver.NewProtocol6WithError(New("test")()), // Assuming New("test")() returns the correct factory type.
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables for acceptance tests.
	requiredEnvVars := []string{
		"UPTIMEKUMA_BASE_URL",
		"UPTIMEKUMA_USERNAME",
		"UPTIMEKUMA_PASSWORD",
	}

	for _, env := range requiredEnvVars {
		if v := os.Getenv(env); v == "" {
			t.Fatalf("%s environment variable must be set for acceptance tests.", env)
		}
	}
}
