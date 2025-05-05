package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestStatusPageOperations tests status page API operations.
func TestStatusPageOperations(t *testing.T) {
	statusPages := []*StatusPage{
		{
			ID:             1,
			Slug:           "main-status",
			Title:          "Main Status Page",
			Description:    "Main system status",
			Theme:          "light",
			Published:      true,
			ShowTags:       false,
			DomainNameList: []string{"status.example.com"},
			Icon:           "/icon.svg",
			ShowPoweredBy:  true,
			PublicGroupList: []PublicGroup{
				{
					ID:          1,
					Name:        "API Services",
					Weight:      1,
					MonitorList: []int{1, 2, 3},
				},
			},
		},
		{
			ID:             2,
			Slug:           "dev-status",
			Title:          "Development Status",
			Description:    "Development system status",
			Theme:          "dark",
			Published:      false,
			ShowTags:       true,
			DomainNameList: []string{"dev-status.example.com"},
			Icon:           "/dev-icon.svg",
			ShowPoweredBy:  false,
			PublicGroupList: []PublicGroup{
				{
					ID:          2,
					Name:        "Dev Services",
					Weight:      1,
					MonitorList: []int{4, 5},
				},
			},
		},
	}

	// Setup mock server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle authentication.
		if r.URL.Path == "/login/access-token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(TokenResponse{
				AccessToken: "test-token-12345",
				TokenType:   "Bearer",
			})
			if err != nil {
				fmt.Printf("ERROR encoding token response: %v\n", err)
				http.Error(w, "failed to encode token response", http.StatusInternalServerError)
			}
			return
		}

		// Check auth header.
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token-12345" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Handle status page requests.
		if r.URL.Path == "/status-pages" {
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				statusPagesValues := make([]StatusPage, 0, len(statusPages))
				for _, spPtr := range statusPages {
					if spPtr != nil {
						statusPagesValues = append(statusPagesValues, *spPtr)
					}
				}

				err := json.NewEncoder(w).Encode(StatusPageList{
					StatusPages: statusPagesValues,
				})

				if err != nil {
					fmt.Printf("ERROR encoding status page list: %v\n", err)
					http.Error(w, "failed to encode status page list", http.StatusInternalServerError)
				}
				return
			case http.MethodPost:
				// Create status page.
				var request AddStatusPageRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					http.Error(w, "Bad request body", http.StatusBadRequest)
					return
				}

				// Check for duplicate slug.
				for _, sp := range statusPages {
					if sp.Slug == request.Slug {
						http.Error(w, "Duplicate slug", http.StatusConflict)
						return
					}
				}

				// Create new status page.
				newStatusPage := &StatusPage{
					ID:             len(statusPages) + 1, // Simple ID assignment for test.
					Slug:           request.Slug,
					Title:          request.Title,
					Description:    request.Msg, // Assuming Msg maps to Description.
					Theme:          "light",     // Default or from request if included.
					Published:      true,        // Default or from request.
					ShowTags:       false,       // Default or from request.
					DomainNameList: []string{},
					Icon:           "/icon.svg", // Default or from request.
					ShowPoweredBy:  true,        // Default or from request.
				}
				statusPages = append(statusPages, newStatusPage)

				w.WriteHeader(http.StatusOK) // Or http.StatusCreated (201).
				err := json.NewEncoder(w).Encode(AddStatusPageResponse{
					Msg: "Status page created",
					// Optionally return the ID or slug if the real API does.
				})
				if err != nil {
					fmt.Printf("ERROR encoding add status page response: %v\n", err)
					http.Error(w, "failed to encode add status page response", http.StatusInternalServerError)
				}
				return
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
		} else if strings.HasPrefix(r.URL.Path, "/status-pages/") {
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(parts) < 2 || parts[0] != "status-pages" {
				http.Error(w, "Bad request path", http.StatusBadRequest)
				return
			}

			slug := parts[1]

			// Handle special endpoints like /incident or /unpin.
			if len(parts) > 2 {
				action := parts[2]
				if action == "incident" {
					if len(parts) > 3 && parts[3] == "unpin" {
						// Unpin incident.
						if r.Method == http.MethodDelete { // Or POST/PUT depending on API.
							w.WriteHeader(http.StatusOK)
							err := json.NewEncoder(w).Encode(UnpinIncidentResponse{
								Detail: "Incident unpinned",
							})
							if err != nil {
								fmt.Printf("ERROR encoding unpin incident response: %v\n", err)
								http.Error(w, "failed to encode unpin incident response", http.StatusInternalServerError)
							}
							return
						}
						http.Error(w, "Method not allowed for unpin", http.StatusMethodNotAllowed)
						return
					} else {
						// Post incident.
						if r.Method == http.MethodPost {
							var request PostIncidentRequest
							if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
								http.Error(w, "Bad request body for incident", http.StatusBadRequest)
								return
							}
							w.WriteHeader(http.StatusOK) // Or 201 Created.
							err := json.NewEncoder(w).Encode(PostIncidentResponse{
								// Simulate response, ID might come from DB in real API.
								ID:          123, // Example incident ID.
								Title:       request.Title,
								Content:     request.Content,
								Style:       request.Style,
								CreatedDate: time.Now().Format(time.RFC3339),
								Pin:         true, // Assuming default or based on request.
							})
							if err != nil {
								fmt.Printf("ERROR encoding post incident response: %v\n", err)
								http.Error(w, "failed to encode post incident response", http.StatusInternalServerError)
							}
							return
						}
						http.Error(w, "Method not allowed for post incident", http.StatusMethodNotAllowed)
						return
					}
				}
				// Add other potential actions here if needed.
				http.Error(w, "Invalid action for status page", http.StatusBadRequest)
				return
			}

			// If not an action endpoint, handle CRUD for the status page itself.

			// Find status page by slug.
			var spIndex = -1
			for i, sp := range statusPages {
				if sp.Slug == slug {
					spIndex = i
					break
				}
			}

			if spIndex == -1 {
				http.NotFound(w, r)
				return
			}

			switch r.Method {
			case http.MethodGet:
				// Get status page.
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(statusPages[spIndex])
				if err != nil {
					fmt.Printf("ERROR encoding single status page: %v\n", err)
					http.Error(w, "failed to encode single status page", http.StatusInternalServerError)
				}
				return
			case http.MethodPost: // Assuming POST for update based on original code, PUT/PATCH might be more standard.
				// Update status page
				var request SaveStatusPageRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					http.Error(w, "Bad request body for update", http.StatusBadRequest)
					return
				}
				currentPage := statusPages[spIndex] // Get pointer to modify in place.

				if request.Title != "" {
					currentPage.Title = request.Title
				}
				if request.Description != "" {
					currentPage.Description = request.Description
				}
				if request.Theme != "" {
					currentPage.Theme = request.Theme
				}
				currentPage.Published = request.Published // Assuming bool always provided.
				currentPage.ShowTags = request.ShowTags   // Assuming bool always provided.
				if request.DomainNameList != nil {        // Check if field exists in request.
					currentPage.DomainNameList = request.DomainNameList
				}
				if request.FooterText != "" {
					currentPage.FooterText = request.FooterText
				}
				if request.CustomCSS != "" {
					currentPage.CustomCSS = request.CustomCSS
				}
				if request.GoogleAnalyticsID != "" {
					currentPage.GoogleAnalyticsID = request.GoogleAnalyticsID
				}
				if request.Icon != "" {
					currentPage.Icon = request.Icon
				}
				currentPage.ShowPoweredBy = request.ShowPoweredBy // Assuming bool always provided.
				if request.PublicGroupList != nil {
					currentPage.PublicGroupList = request.PublicGroupList
				}

				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(SaveStatusPageResponse{
					Detail: "Status page updated",
				})
				if err != nil {
					fmt.Printf("ERROR encoding save status page response: %v\n", err)
					http.Error(w, "failed to encode save status page response", http.StatusInternalServerError)
				}
				return
			case http.MethodDelete:
				// Delete status page.
				statusPages = append(statusPages[:spIndex], statusPages[spIndex+1:]...)
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(DeleteStatusPageResponse{
					Detail: "Status page deleted",
				})
				if err != nil {
					fmt.Printf("ERROR encoding delete status page response: %v\n", err)
					http.Error(w, "failed to encode delete status page response", http.StatusInternalServerError)
				}
				return
			default:
				http.Error(w, "Method not allowed for this resource", http.StatusMethodNotAllowed)
				return
			}
		}

		// If we get here, the path wasn't matched.
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Create client.
	config := &Config{
		BaseURL:  server.URL,
		Username: "testuser",
		Password: "testpass",
		Timeout:  5 * time.Second,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test GetStatusPages.
	initialPages, err := client.GetStatusPages(ctx)
	if err != nil {
		t.Fatalf("GetStatusPages failed: %v", err)
	}
	if len(initialPages) != len(statusPages) {
		t.Errorf("GetStatusPages initial check: Expected %d status pages, got %d", len(statusPages), len(initialPages))
	}

	// Test GetStatusPage.
	page, err := client.GetStatusPage(ctx, "main-status")
	if err != nil {
		t.Fatalf("GetStatusPage failed for 'main-status': %v", err)
	}
	if page.Slug != "main-status" { // Basic check.
		t.Errorf("GetStatusPage returned unexpected slug: %+v", page)
	}

	// Test CreateStatusPage.
	newPage := &AddStatusPageRequest{
		Slug:  "new-status",
		Title: "New Status Page",
		Msg:   "A new status page",
	}
	createResult, err := client.CreateStatusPage(ctx, newPage)
	if err != nil {
		t.Fatalf("CreateStatusPage failed: %v", err)
	}
	if !strings.Contains(createResult.Msg, "created") { // Check message content.
		t.Errorf("CreateStatusPage returned unexpected result message: %+v", createResult)
	}
	// Verify creation by trying to get the new page.
	createdPage, err := client.GetStatusPage(ctx, "new-status")
	if err != nil {
		t.Fatalf("GetStatusPage failed for newly created 'new-status': %v", err)
	}
	if createdPage.Title != "New Status Page" {
		t.Errorf("Newly created status page has wrong title: %+v", createdPage)
	}

	// Test UpdateStatusPage.
	updatePageSlug := "main-status"
	updatePageData := &SaveStatusPageRequest{
		Title:       "Updated Main Status Page",
		Description: "Updated description for main status",
		Theme:       "dark",
		Published:   true,
	}
	updateResult, err := client.UpdateStatusPage(ctx, updatePageSlug, updatePageData)
	if err != nil {
		t.Fatalf("UpdateStatusPage failed for '%s': %v", updatePageSlug, err)
	}

	// Assert that the Detail field (which is interface{}) holds a string.
	detailStr, ok := updateResult.Detail.(string)
	if !ok {
		// If the underlying type isn't a string, fail the test.
		t.Fatalf("UpdateStatusPage response field 'Detail' was not a string: type=%T, value=%+v",
			updateResult.Detail, updateResult)
	}

	if !strings.Contains(detailStr, "updated") {
		t.Errorf("UpdateStatusPage returned unexpected result detail string: '%s'", detailStr)
	}

	// Verify update by getting the page again.
	updatedPage, err := client.GetStatusPage(ctx, updatePageSlug)
	if err != nil {
		t.Fatalf("GetStatusPage failed for updated '%s': %v", updatePageSlug, err)
	}
	if updatedPage.Title != "Updated Main Status Page" || updatedPage.Theme != "dark" {
		t.Errorf("UpdateStatusPage did not apply changes correctly: %+v", updatedPage)
	}

	// Test PostIncident.
	incidentSlug := "main-status" // Post to existing page.
	incidentData := &PostIncidentRequest{
		Title:   "Test Incident",
		Content: "There is a problem with the service",
		Style:   "warning",
	}
	incidentResult, err := client.PostIncident(ctx, incidentSlug, incidentData)
	if err != nil {
		t.Fatalf("PostIncident failed for '%s': %v", incidentSlug, err)
	}
	if incidentResult.Title != "Test Incident" { // Basic check on response.
		t.Errorf("PostIncident returned unexpected title in result: %+v", incidentResult)
	}

	// Test UnpinIncident.
	unpinSlug := "main-status" // Unpin from existing page.
	unpinResult, err := client.UnpinIncident(ctx, unpinSlug)
	if err != nil {
		t.Fatalf("UnpinIncident failed for '%s': %v", unpinSlug, err)
	}
	if !strings.Contains(unpinResult.Detail, "unpinned") {
		t.Errorf("UnpinIncident returned unexpected result detail: %+v", unpinResult)
	}

	// Test DeleteStatusPage.
	deleteSlug := "new-status" // Delete the one we created.
	deleteResult, err := client.DeleteStatusPage(ctx, deleteSlug)
	if err != nil {
		t.Fatalf("DeleteStatusPage failed for '%s': %v", deleteSlug, err)
	}
	if !strings.Contains(deleteResult.Detail, "deleted") {
		t.Errorf("DeleteStatusPage returned unexpected result detail: %+v", deleteResult)
	}
	// Verify deletion.
	_, err = client.GetStatusPage(ctx, deleteSlug)
	if err == nil {
		t.Errorf("GetStatusPage should have failed for deleted slug '%s', but succeeded", deleteSlug)
	} else {
		fmt.Printf("DEBUG: Verified status page '%s' deletion (expected error): %v\n", deleteSlug, err)
	}

}

