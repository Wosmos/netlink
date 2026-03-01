package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"netlink/auth"
	"netlink/middleware"
	"netlink/models"
	"netlink/repository"
)

type NoteHandler struct {
	repo        *repository.NoteRepository
	authService *auth.AuthService
}

func NewNoteHandler(repo *repository.NoteRepository, authService *auth.AuthService) *NoteHandler {
	return &NoteHandler{repo: repo, authService: authService}
}

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Color   string `json:"color"`
}

type UpdateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Color   string `json:"color"`
	Pinned  bool   `json:"pinned"`
}

func (h *NoteHandler) requireAuth(w http.ResponseWriter, r *http.Request) int {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		middleware.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return 0
	}
	return user.ID
}

// GET /api/notes
func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	convIDStr := r.URL.Query().Get("conversation_id")
	var notes []models.Note
	var err error

	if convIDStr != "" {
		convID, parseErr := strconv.Atoi(convIDStr)
		if parseErr != nil {
			middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
			return
		}
		notes, err = h.repo.GetByConversation(convID)
	} else {
		notes, err = h.repo.GetAllByUser(userID)
	}

	if err != nil {
		middleware.JSONError(w, "Failed to get notes", http.StatusInternalServerError)
		return
	}

	if notes == nil {
		notes = []models.Note{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": notes})
}

// POST /api/notes
func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	var req struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		Color          string `json:"color"`
		ConversationID *int   `json:"conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Title) > 500 {
		middleware.JSONError(w, "Title too long (max 500 chars)", http.StatusBadRequest)
		return
	}
	if len(req.Content) > 50000 {
		middleware.JSONError(w, "Content too long (max 50000 chars)", http.StatusBadRequest)
		return
	}

	if req.Color == "" {
		req.Color = "#ffffff"
	}

	var note *models.Note
	var err error
	if req.ConversationID != nil && *req.ConversationID > 0 {
		note, err = h.repo.CreateForConversation(userID, *req.ConversationID, req.Title, req.Content, req.Color)
	} else {
		note, err = h.repo.Create(userID, req.Title, req.Content, req.Color)
	}

	if err != nil {
		middleware.JSONError(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": note})
}

// GET /api/notes/{id}
func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.JSONError(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	note, err := h.repo.GetByID(id, userID)
	if err != nil {
		middleware.JSONError(w, "Note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": note})
}

// PUT /api/notes/{id}
func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.JSONError(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Title) > 500 {
		middleware.JSONError(w, "Title too long (max 500 chars)", http.StatusBadRequest)
		return
	}
	if len(req.Content) > 50000 {
		middleware.JSONError(w, "Content too long (max 50000 chars)", http.StatusBadRequest)
		return
	}

	if err := h.repo.Update(id, userID, req.Title, req.Content, req.Color, req.Pinned); err != nil {
		middleware.JSONError(w, "Failed to update note", http.StatusInternalServerError)
		return
	}

	note, err := h.repo.GetByID(id, userID)
	if err != nil {
		middleware.JSONError(w, "Failed to get updated note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": note})
}

// DELETE /api/notes/{id}
func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.JSONError(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(id, userID); err != nil {
		middleware.JSONError(w, "Failed to delete note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// POST /api/notes/{id}/pin
func (h *NoteHandler) TogglePin(w http.ResponseWriter, r *http.Request) {
	userID := h.requireAuth(w, r)
	if userID == 0 {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.JSONError(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.TogglePin(id, userID); err != nil {
		middleware.JSONError(w, "Failed to toggle pin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
