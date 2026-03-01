package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		statusCode int
	}{
		{"bad request", "Invalid input", http.StatusBadRequest},
		{"not found", "Resource not found", http.StatusNotFound},
		{"internal error", "Something went wrong", http.StatusInternalServerError},
		{"unauthorized", "Not authenticated", http.StatusUnauthorized},
		{"forbidden", "Access denied", http.StatusForbidden},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			JSONError(w, tc.message, tc.statusCode)

			if w.Code != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type=%q, got %q", "application/json", contentType)
			}

			var resp map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode JSON: %v", err)
			}

			if resp["success"] != false {
				t.Error("Expected success=false")
			}
			if resp["error"] != tc.message {
				t.Errorf("Expected error=%q, got %q", tc.message, resp["error"])
			}
		})
	}
}

func TestJSONMethodNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	JSONMethodNotAllowed(w)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type=%q, got %q", "application/json", contentType)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if resp["success"] != false {
		t.Error("Expected success=false")
	}
	if resp["error"] != "Method not allowed" {
		t.Errorf("Expected error=%q, got %q", "Method not allowed", resp["error"])
	}
}
