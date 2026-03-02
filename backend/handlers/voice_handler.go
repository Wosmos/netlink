package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"netlink/middleware"
	"netlink/storage"
)

var mimeToExt = map[string]string{
	"audio/webm":             ".webm",
	"audio/webm;codecs=opus": ".webm",
	"audio/mp4":              ".m4a",
	"audio/mpeg":             ".mp3",
	"audio/m4a":              ".m4a",
	"audio/aac":              ".m4a",
	"audio/x-m4a":            ".m4a",
}

var extToMime = map[string]string{
	".webm": "audio/webm",
	".m4a":  "audio/mp4",
	".mp4":  "audio/mp4",
	".mp3":  "audio/mpeg",
}

type VoiceHandler struct {
	repo        ChatRepoInterface
	userRepo    UserRepoInterface
	authService AuthServiceInterface
	storage     *storage.SupabaseStorage
	maxFileSize int64
	maxDuration int
}

func NewVoiceHandler(repo ChatRepoInterface, userRepo UserRepoInterface, authService AuthServiceInterface, store *storage.SupabaseStorage) *VoiceHandler {
	return &VoiceHandler{
		repo:        repo,
		userRepo:    userRepo,
		authService: authService,
		storage:     store,
		maxFileSize: 50 * 1024 * 1024, // 50MB max
		maxDuration: 600,              // 10 minutes max
	}
}

func (h *VoiceHandler) requireAuth(w http.ResponseWriter, r *http.Request) int {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		middleware.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return 0
	}
	return user.ID
}

// POST /api/voice/upload
func (h *VoiceHandler) UploadVoice(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	log.Printf("Voice upload request from user %d", userID)

	if err := r.ParseMultipartForm(h.maxFileSize); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		middleware.JSONError(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		log.Printf("Error getting audio file: %v", err)
		middleware.JSONError(w, "No audio file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received audio file: %s, size: %d bytes", header.Filename, header.Size)

	durationStr := r.FormValue("duration")
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil || duration <= 0 || duration > float64(h.maxDuration) {
		middleware.JSONError(w, "Invalid duration", http.StatusBadRequest)
		return
	}

	waveformStr := r.FormValue("waveform")
	var waveform []float64
	if waveformStr != "" {
		if err := json.Unmarshal([]byte(waveformStr), &waveform); err != nil {
			waveform = nil
		}
	}

	// Determine file extension from content type
	contentType := header.Header.Get("Content-Type")
	baseType := strings.TrimSpace(strings.Split(contentType, ";")[0])

	ext, ok := mimeToExt[contentType]
	if !ok {
		ext, ok = mimeToExt[baseType]
	}
	if !ok {
		if baseType == "application/octet-stream" || baseType == "" {
			ext = ".m4a"
		} else {
			middleware.JSONError(w, "Invalid audio format. Supported: webm, m4a, mp3, mp4", http.StatusBadRequest)
			return
		}
	}

	// Build the upload MIME for Supabase (use base type without codec params)
	uploadMime := baseType
	if uploadMime == "" || uploadMime == "application/octet-stream" {
		if m, ok := extToMime[ext]; ok {
			uploadMime = m
		} else {
			uploadMime = "audio/webm"
		}
	}

	log.Printf("Audio content-type: %s, using extension: %s", contentType, ext)

	// Upload to Supabase Storage
	timestamp := time.Now().Unix()
	objectPath := fmt.Sprintf("voice/%d/%d%s", userID, timestamp, ext)

	ctx := context.Background()
	_, err = h.storage.Upload(ctx, objectPath, file, uploadMime)
	if err != nil {
		log.Printf("Supabase upload error: %v", err)
		middleware.JSONError(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	log.Printf("Uploaded voice to Supabase: %s (%d bytes)", objectPath, header.Size)

	middleware.JSONSuccess(w, map[string]interface{}{
		"file_path": objectPath,
		"file_size": header.Size,
		"duration":  duration,
		"waveform":  waveform,
	})
}

// GET /api/voice/download?path=...
func (h *VoiceHandler) DownloadVoice(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		middleware.JSONError(w, "No file path provided", http.StatusBadRequest)
		return
	}

	// Security: reject directory traversal
	if strings.Contains(filePath, "..") {
		middleware.JSONError(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Generate a signed URL (1 hour expiry)
	ctx := context.Background()
	signedURL, err := h.storage.SignedURL(ctx, filePath, 3600)
	if err != nil {
		log.Printf("Signed URL error for %s: %v", filePath, err)
		middleware.JSONError(w, "File not found", http.StatusNotFound)
		return
	}

	middleware.JSONSuccess(w, map[string]interface{}{
		"url": signedURL,
	})
}

// DELETE /api/voice/delete?path=...
func (h *VoiceHandler) DeleteVoice(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		middleware.JSONError(w, "No file path provided", http.StatusBadRequest)
		return
	}

	// Security: reject directory traversal
	if strings.Contains(filePath, "..") {
		middleware.JSONError(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Verify user owns this voice file (path format: voice/{userID}/...)
	expectedPrefix := fmt.Sprintf("voice/%d/", userID)
	if !strings.HasPrefix(filePath, expectedPrefix) {
		middleware.JSONError(w, "Access denied", http.StatusForbidden)
		return
	}

	ctx := context.Background()
	if err := h.storage.Delete(ctx, filePath); err != nil {
		log.Printf("Supabase delete error for %s: %v", filePath, err)
		middleware.JSONError(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	middleware.JSONOk(w)
}
