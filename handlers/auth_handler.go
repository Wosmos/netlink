package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"go-to-do/auth"
)

type AuthHandler struct {
	authService  *auth.AuthService
	loginTmpl    *template.Template
	registerTmpl *template.Template
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	loginTmpl := template.Must(template.ParseFiles("templates/login.html"))
	registerTmpl := template.Must(template.ParseFiles("templates/register.html"))
	return &AuthHandler{
		authService:  authService,
		loginTmpl:    loginTmpl,
		registerTmpl: registerTmpl,
	}
}

type AuthPageData struct {
	Error string
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if user, _ := h.authService.GetUserFromRequest(r); user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		h.loginTmpl.Execute(w, nil)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	session, err := h.authService.Login(email, password)
	if err != nil {
		log.Printf("Login error: %v", err)
		h.loginTmpl.Execute(w, AuthPageData{Error: "An error occurred"})
		return
	}
	if session == nil {
		h.loginTmpl.Execute(w, AuthPageData{Error: "Invalid email or password"})
		return
	}

	h.authService.SetSessionCookie(w, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if user, _ := h.authService.GetUserFromRequest(r); user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		h.registerTmpl.Execute(w, nil)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validation
	if !h.authService.ValidateEmail(email) {
		h.registerTmpl.Execute(w, AuthPageData{Error: "Invalid email address"})
		return
	}
	if !h.authService.ValidatePassword(password) {
		h.registerTmpl.Execute(w, AuthPageData{Error: "Password must be 8-128 characters"})
		return
	}
	if password != confirmPassword {
		h.registerTmpl.Execute(w, AuthPageData{Error: "Passwords do not match"})
		return
	}

	if err := h.authService.Register(email, password); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			h.registerTmpl.Execute(w, AuthPageData{Error: "Email already registered"})
			return
		}
		log.Printf("Register error: %v", err)
		h.registerTmpl.Execute(w, AuthPageData{Error: "An error occurred"})
		return
	}

	// Auto-login after registration
	session, err := h.authService.Login(email, password)
	if err != nil || session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.authService.SetSessionCookie(w, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		h.authService.Logout(cookie.Value)
	}
	h.authService.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
