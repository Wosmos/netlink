package handlers

import (
	"go-to-do/auth"
	"go-to-do/models"
	"go-to-do/repository"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
	if err != nil {
		log.Printf("Auth error: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil
	}
	if user == nil {
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

	tasks, err := h.repo.GetAllByUser(user.ID)
	if err != nil {
		http.Error(w, "Error retrieving tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	h.tmpl.Execute(w, HomePageData{User: user, Tasks: tasks})
}

func (h *TaskHandler) Add(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if text := r.FormValue("task"); text != "" {
		if err := h.repo.Create(user.ID, text); err != nil {
			log.Printf("Error adding task: %v", err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TaskHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := h.repo.Toggle(id, user.ID); err != nil {
		log.Printf("Error toggling task: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := h.requireAuth(w, r)
	if user == nil {
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := h.repo.Delete(id, user.ID); err != nil {
		log.Printf("Error deleting task: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
