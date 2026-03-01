package middleware

import (
	"encoding/json"
	"io"
	"net/http"
)

// JSONError writes a consistent JSON error response.
func JSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// JSONSuccess writes a consistent JSON success response with optional data.
func JSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{"success": true}
	if data != nil {
		resp["data"] = data
	}
	json.NewEncoder(w).Encode(resp)
}

// JSONSuccessMessage writes a success response with a message field.
func JSONSuccessMessage(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
	})
}

// JSONCreated writes a 201 success response with data.
func JSONCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := map[string]interface{}{"success": true}
	if data != nil {
		resp["data"] = data
	}
	json.NewEncoder(w).Encode(resp)
}

// JSONMethodNotAllowed writes a 405 JSON response.
func JSONMethodNotAllowed(w http.ResponseWriter) {
	JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// MaxBodyReader wraps r.Body with a size-limited reader to prevent memory exhaustion.
// limit is in bytes. Use 0 for no limit.
func MaxBodyReader(r *http.Request, limit int64) {
	r.Body = http.MaxBytesReader(nil, r.Body, limit)
}

// LimitedBody returns a reader limited to maxBytes. Attach to r.Body before decoding.
func LimitedBody(body io.ReadCloser, maxBytes int64) io.ReadCloser {
	return http.MaxBytesReader(nil, body, maxBytes)
}

// LimitBody wraps a handler with a request body size limit.
// Use this on JSON API endpoints to prevent memory exhaustion from oversized payloads.
func LimitBody(maxBytes int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next(w, r)
	}
}