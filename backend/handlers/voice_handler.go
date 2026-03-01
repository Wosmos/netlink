package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"netlink/auth"
	"netlink/repository"
)

// Maps content types to file extensions and back
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
	repo        *repository.ChatRepository
	userRepo    *repository.UserRepository
	authService *auth.AuthService
	uploadDir   string
	maxFileSize int64
	maxDuration int
}

func NewVoiceHandler(repo *repository.ChatRepository, userRepo *repository.UserRepository, authService *auth.AuthService) *VoiceHandler {
	uploadDir := os.Getenv("VOICE_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads/voice"
	}

	// Create upload directory if it doesn't exist
	// NOTE: For production, use object storage (S3, MinIO, Backblaze B2) instead of local filesystem
	// This ensures scalability, redundancy, and CDN integration for global delivery
	// Example: AWS S3, Google Cloud Storage, Azure Blob Storage, or self-hosted MinIO
	os.MkdirAll(uploadDir, 0755)

	return &VoiceHandler{
		repo:        repo,
		userRepo:    userRepo,
		authService: authService,
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024, // 50MB max
		maxDuration: 600,              // 10 minutes max
	}
}

func (h *VoiceHandler) requireAuth(w http.ResponseWriter, r *http.Request) int {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
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

	// Parse multipart form
	if err := r.ParseMultipartForm(h.maxFileSize); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "File too large or invalid form data"})
		return
	}

	// Get file from form
	file, header, err := r.FormFile("audio")
	if err != nil {
		log.Printf("Error getting audio file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No audio file provided"})
		return
	}
	defer file.Close()

	log.Printf("Received audio file: %s, size: %d bytes", header.Filename, header.Size)

	// Get metadata
	durationStr := r.FormValue("duration")
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil || duration <= 0 || duration > float64(h.maxDuration) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid duration"})
		return
	}

	// Get waveform data (optional)
	waveformStr := r.FormValue("waveform")
	var waveform []float64
	if waveformStr != "" {
		if err := json.Unmarshal([]byte(waveformStr), &waveform); err != nil {
			// Ignore waveform if invalid
			waveform = nil
		}
	}

	// Validate file type and determine extension
	contentType := header.Header.Get("Content-Type")
	// Normalize: strip parameters like charset but keep codecs for webm
	baseType := strings.TrimSpace(strings.Split(contentType, ";")[0])

	ext, ok := mimeToExt[contentType]
	if !ok {
		ext, ok = mimeToExt[baseType]
	}
	if !ok {
		// Try to infer from filename extension (handles application/octet-stream, etc.)
		origExt := strings.ToLower(filepath.Ext(header.Filename))
		if _, known := extToMime[origExt]; known {
			ext = origExt
		} else if baseType == "application/octet-stream" || baseType == "" {
			// Mobile apps often send octet-stream; default to m4a (most common mobile format)
			ext = ".m4a"
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid audio format. Supported: webm, m4a, mp3, mp4"})
			return
		}
	}

	log.Printf("Audio content-type: %s, using extension: %s", contentType, ext)

	// Generate unique filename with correct extension
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d%s", userID, timestamp, ext)

	// Create user directory
	userDir := filepath.Join(h.uploadDir, strconv.Itoa(userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to create directory"})
		return
	}

	// Save file
	filePath := filepath.Join(userDir, filename)

	// NOTE: For production, replace local file storage with object storage:
	// - AWS S3: Use aws-sdk-go-v2 to upload directly to S3 bucket
	// - MinIO: Self-hosted S3-compatible storage (FREE, recommended)
	// - Backblaze B2: Cost-effective alternative to S3
	// - Google Cloud Storage: Good for global distribution
	//
	// Benefits of object storage:
	// - Scalability: Handle millions of files
	// - Redundancy: Built-in backup and replication
	// - CDN: Fast global delivery with CloudFlare/CloudFront
	// - Cost: Pay only for what you use
	// - Security: Built-in encryption and access control
	//
	// Example MinIO integration:
	// minioClient.PutObject(ctx, "voice-messages", filename, file, header.Size, minio.PutObjectOptions{})

	dst, err := os.Create(filePath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to save file"})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to save file"})
		return
	}

	// Return file info
	relativePath := filepath.Join(strconv.Itoa(userID), filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"file_path": relativePath,
			"file_size": written,
			"duration":  duration,
			"waveform":  waveform,
		},
	})
}

// GET /api/voice/download/:path
func (h *VoiceHandler) DownloadVoice(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	// Get file path from query
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No file path provided"})
		return
	}

	// Security: Prevent directory traversal
	cleanPath := filepath.Clean(filePath)
	if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid file path"})
		return
	}

	fullPath := filepath.Join(h.uploadDir, cleanPath)

	// Double-check: resolved path must be inside uploadDir
	absUpload, _ := filepath.Abs(h.uploadDir)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absUpload+string(filepath.Separator)) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid file path"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "File not found"})
		return
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to open file"})
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to get file info"})
		return
	}

	// Determine correct MIME type from file extension
	fileExt := strings.ToLower(filepath.Ext(fullPath))
	serveMime, ok := extToMime[fileExt]
	if !ok {
		serveMime = "audio/webm" // fallback
	}

	// Set headers for audio streaming
	w.Header().Set("Content-Type", serveMime)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Support range requests for seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		// Parse range header (simplified - production should use proper range parsing)
		var start, end int64
		fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

		if end == 0 || end >= fileInfo.Size() {
			end = fileInfo.Size() - 1
		}

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
		w.WriteHeader(http.StatusPartialContent)

		file.Seek(start, 0)
		io.CopyN(w, file, end-start+1)
	} else {
		// Stream entire file
		io.Copy(w, file)
	}
}

// DELETE /api/voice/delete
func (h *VoiceHandler) DeleteVoice(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No file path provided"})
		return
	}

	// Security: Prevent directory traversal and ensure user owns the file
	cleanPath := filepath.Clean(filePath)
	if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid file path"})
		return
	}

	userPrefix := strconv.Itoa(userID) + string(filepath.Separator)
	if !strings.HasPrefix(cleanPath, userPrefix) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Access denied"})
		return
	}

	fullPath := filepath.Join(h.uploadDir, cleanPath)

	// Double-check: resolved path must be inside uploadDir
	absUpload, _ := filepath.Abs(h.uploadDir)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absUpload+string(filepath.Separator)) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid file path"})
		return
	}

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "File not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to delete file"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
