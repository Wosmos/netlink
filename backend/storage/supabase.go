package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SupabaseStorage struct {
	projectURL string
	serviceKey string
	bucket     string
	client     *http.Client
}

func NewSupabaseStorage(projectURL, serviceKey, bucket string) *SupabaseStorage {
	return &SupabaseStorage{
		projectURL: strings.TrimRight(projectURL, "/"),
		serviceKey: serviceKey,
		bucket:     bucket,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Upload uploads a file to Supabase Storage and returns the object path.
func (s *SupabaseStorage) Upload(ctx context.Context, objectPath string, data io.Reader, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, s.bucket, objectPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, data)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	return objectPath, nil
}

// SignedURL generates a temporary signed URL for a private object.
func (s *SupabaseStorage) SignedURL(ctx context.Context, objectPath string, expiresIn int) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", s.projectURL, s.bucket, objectPath)

	payload := fmt.Sprintf(`{"expiresIn":%d}`, expiresIn)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sign request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sign failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	// Supabase returns a relative path like "/object/sign/bucket/path?token=xxx"
	// We need to prepend the project URL + "/storage/v1"
	if strings.HasPrefix(result.SignedURL, "/") {
		return s.projectURL + "/storage/v1" + result.SignedURL, nil
	}
	return result.SignedURL, nil
}

// Delete removes a file from Supabase Storage.
func (s *SupabaseStorage) Delete(ctx context.Context, objectPath string) error {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, s.bucket, objectPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
