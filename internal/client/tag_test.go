package client

import (
	"context"
	"encoding/json"
	"fmt" // Import fmt for logging errors in handler.
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestTagOperations tests tag API operations.
func TestTagOperations(t *testing.T) {
	// Setup tags for the mock server (use pointers if modifying in handler).
	tags := []*Tag{
		{
			ID:    1,
			Name:  "production",
			Color: "#00FF00",
		},
		{
			ID:    2,
			Name:  "development",
			Color: "#0000FF",
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

		// Handle tag requests.
		if r.URL.Path == "/tags" {
			switch r.Method {
			case http.MethodGet:
				// List tags.
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(tags)
				if err != nil {
					fmt.Printf("ERROR encoding tag list: %v\n", err)
					http.Error(w, "failed to encode tag list", http.StatusInternalServerError)
				}
				return
			case http.MethodPost:
				// Create tag
				var newTag Tag // Use value type here, ID will be assigned.
				if err := json.NewDecoder(r.Body).Decode(&newTag); err != nil {
					http.Error(w, "Bad request body", http.StatusBadRequest)
					return
				}
				newTag.ID = len(tags) + 1    // Simple ID assignment for test.
				tags = append(tags, &newTag) // Append pointer if tags slice holds pointers.
				w.WriteHeader(http.StatusOK) // Or http.StatusCreated (201).
				// Return the tag with the assigned ID.
				err := json.NewEncoder(w).Encode(newTag)
				if err != nil {
					fmt.Printf("ERROR encoding created tag: %v\n", err)
					http.Error(w, "failed to encode created tag", http.StatusInternalServerError)
				}
				return
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
		} else if strings.HasPrefix(r.URL.Path, "/tags/") {
			// Extract tag ID from path.
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(parts) < 2 || parts[0] != "tags" {
				http.Error(w, "Bad request path", http.StatusBadRequest)
				return
			}

			idStr := parts[1]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid tag ID format", http.StatusBadRequest)
				return
			}

			// Find tag by ID.
			var tagIndex = -1
			for i, tg := range tags {
				if tg.ID == id {
					tagIndex = i
					break
				}
			}

			if tagIndex == -1 {
				http.NotFound(w, r)
				return
			}

			switch r.Method {
			case http.MethodGet:
				// Get tag.
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(tags[tagIndex])
				if err != nil {
					fmt.Printf("ERROR encoding single tag: %v\n", err)
					http.Error(w, "failed to encode single tag", http.StatusInternalServerError)
				}
				return
			case http.MethodDelete:
				// Delete tag.
				tags = append(tags[:tagIndex], tags[tagIndex+1:]...)
				w.WriteHeader(http.StatusOK)
				// Optionally encode a success message.
				return
			// Add PUT/PATCH for tag updates if needed.
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

	// --- Test Operations ---.

	// Test GetTags.
	initialTags, err := client.GetTags(ctx) // Renamed variable
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}
	// Use reflect.DeepEqual for slice comparison if appropriate, otherwise check length/key fields.
	if len(initialTags) != len(tags) {
		t.Errorf("GetTags initial check: Expected %d tags, got %d", len(tags), len(initialTags))
	}
	// Add more specific checks if DeepEqual is problematic with pointers/state.

	// Test GetTag.
	tag, err := client.GetTag(ctx, 1)
	if err != nil {
		t.Fatalf("GetTag failed for ID 1: %v", err)
	}
	if tag.ID != 1 || tag.Name != "production" || tag.Color != "#00FF00" {
		t.Errorf("GetTag returned unexpected result: %+v", tag)
	}

	// Test CreateTag.
	newTagData := &Tag{
		Name:  "testing",
		Color: "#FF0000",
	}
	createdTag, err := client.CreateTag(ctx, newTagData)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}
	if createdTag.Name != newTagData.Name || createdTag.Color != newTagData.Color {
		t.Errorf("CreateTag returned unexpected data:\nExpected (partial): %+v\nGot:              %+v", newTagData, createdTag)
	}
	if createdTag.ID == 0 {
		t.Errorf("CreateTag did not assign an ID: %+v", createdTag)
	}

	// Test DeleteTag.
	deleteTargetID := 2 // Delete the 'development' tag.
	if err := client.DeleteTag(ctx, deleteTargetID); err != nil {
		t.Fatalf("DeleteTag failed for ID %d: %v", deleteTargetID, err)
	}

	// Verify tag was deleted.
	finalTags, err := client.GetTags(ctx) // Renamed variable.
	if err != nil {
		t.Fatalf("GetTags failed after deletion: %v", err)
	}

	// Should have 2 tags now (ID 1 'production' and the newly created one).
	expectedTagCount := 2
	if len(finalTags) != expectedTagCount {
		t.Errorf("Expected %d tags after deletion, got %d", expectedTagCount, len(finalTags))
	}

	// Verify the deleted tag (ID 2) is not in the list.
	foundDeleted := false
	for _, tg := range finalTags {
		if tg.ID == deleteTargetID {
			foundDeleted = true
			break
		}
	}
	if foundDeleted {
		t.Errorf("Tag ID %d still exists after deletion", deleteTargetID)
	} else {
		fmt.Printf("DEBUG: Verified tag %d deletion.\n", deleteTargetID)
	}
}
