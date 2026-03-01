package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"netlink/auth"
	"netlink/middleware"
	"netlink/models"
	"netlink/repository"
)

type TaskHandler struct {
	repo        *repository.TaskRepository
	authService *auth.AuthService
	tmpl        *template.Template
}

type HomePageData struct {
	User  *models.User
	Tasks []models.Task
}

func NewTaskHandler(repo *repository.TaskRepository, authService *auth.AuthService) *TaskHandler {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	return &TaskHandler{repo: repo, authService: authService, tmpl: tmpl}
}

func (h *TaskHandler) requireAuth(w http.ResponseWriter, r *http.Request) *models.User {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil
	}
	return user
}

func (h *TaskHandler) Home(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}
	tasks, _ := h.repo.GetAllByUser(user.ID)
	h.tmpl.Execute(w, HomePageData{User: user, Tasks: tasks})
}

func (h *TaskHandler) Add(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil || r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if text := r.FormValue("task"); text != "" {
		h.repo.Create(user.ID, text)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TaskHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil || r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
		h.repo.Toggle(id, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil || r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
		h.repo.Delete(id, user.ID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ============ API Methods ============

func (h *TaskHandler) requireAuthAPI(w http.ResponseWriter, r *http.Request) *models.User {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		middleware.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}
	return user
}

func (h *TaskHandler) APIList(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}

	// Check if conversation_id is provided
	convIDStr := r.URL.Query().Get("conversation_id")
	var tasks []models.Task
	var err error

	if convIDStr != "" {
		convID, parseErr := strconv.Atoi(convIDStr)
		if parseErr != nil {
			middleware.JSONError(w, "Invalid conversation ID", http.StatusBadRequest)
			return
		}
		tasks, err = h.repo.GetByConversation(convID, user.ID)
	} else {
		tasks, err = h.repo.GetAllByUser(user.ID)
	}

	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		tasks = []models.Task{}
	}
	middleware.JSONSuccess(w, tasks)
}

func (h *TaskHandler) APICreate(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	var req struct {
		Text           string `json:"text"`
		ConversationID *int   `json:"conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" || len(req.Text) > 5000 {
		middleware.JSONError(w, "Text required (max 5000 chars)", http.StatusBadRequest)
		return
	}

	var err error
	if req.ConversationID != nil && *req.ConversationID > 0 {
		err = h.repo.CreateForConversation(user.ID, *req.ConversationID, req.Text)
	} else {
		err = h.repo.Create(user.ID, req.Text)
	}

	if err != nil {
		middleware.JSONError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	middleware.JSONCreated(w, nil)
}

func (h *TaskHandler) APIToggle(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	h.repo.Toggle(id, user.ID)
	middleware.JSONOk(w)
}

func (h *TaskHandler) APIDelete(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	h.repo.Delete(id, user.ID)
	middleware.JSONOk(w)
}
