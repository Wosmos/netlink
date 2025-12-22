package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"go-to-do/auth"
	"go-to-do/models"
	"go-to-do/repository"
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return nil
	}
	return user
}

func (h *TaskHandler) APIList(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	tasks, err := h.repo.GetAllByUser(user.ID)
	if err != nil {
		log.Printf("Error getting tasks: %v", err)
		tasks = []models.Task{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": tasks})
}

func (h *TaskHandler) APICreate(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Text required"})
		return
	}
	h.repo.Create(user.ID, req.Text)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (h *TaskHandler) APIToggle(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	h.repo.Toggle(id, user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (h *TaskHandler) APIDelete(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuthAPI(w, r)
	if user == nil {
		return
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	h.repo.Delete(id, user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
