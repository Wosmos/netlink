package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"go-to-do/auth"
)

type AuthHandler struct {
	authService        *auth.AuthService
	loginTmpl          *template.Template
	registerTmpl       *template.Template
	forgotPasswordTmpl *template.Template
	resetPasswordTmpl  *template.Template
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	loginTmpl := template.Must(template.ParseFiles("templates/login.html"))
	registerTmpl := template.Must(template.ParseFiles("templates/register.html"))
	forgotPasswordTmpl := template.Must(template.ParseFiles("templates/forgot_password.html"))
	resetPasswordTmpl := template.Must(template.ParseFiles("templates/reset_password.html"))
	return &AuthHandler{
		authService:        authService,
		loginTmpl:          loginTmpl,
		registerTmpl:       registerTmpl,
		forgotPasswordTmpl: forgotPasswordTmpl,
		resetPasswordTmpl:  resetPasswordTmpl,
	}
}

type AuthPageData struct {
	Error   string
	Success string
	Token   string
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if user, _ := h.authService.GetUserFromRequest(r); user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		h.loginTmpl.Execute(w, AuthPageData{
			Success: r.URL.Query().Get("success"),
			Error:   r.URL.Query().Get("error"),
		})
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	session, err := h.authService.Login(email, password)
	if err != nil {
		h.loginTmpl.Execute(w, AuthPageData{Error: err.Error()})
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

	// Redirect to login with success message instead of auto-login
	http.Redirect(w, r, "/login?success=Verification email sent! Please check your inbox.", http.StatusSeeOther)
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/login?error=Invalid verification link", http.StatusSeeOther)
		return
	}

	if err := h.authService.VerifyEmail(token); err != nil {
		http.Redirect(w, r, "/login?error=Verification failed: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/login?success=Email verified! You can now log in.", http.StatusSeeOther)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.forgotPasswordTmpl.Execute(w, nil)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	if err := h.authService.ForgotPassword(email); err != nil {
		log.Printf("Forgot password error: %v", err)
		h.forgotPasswordTmpl.Execute(w, AuthPageData{Error: "An error occurred"})
		return
	}

	h.forgotPasswordTmpl.Execute(w, AuthPageData{Success: "If that email is registered, you will receive a reset link."})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.FormValue("token")
	}

	if token == "" {
		http.Redirect(w, r, "/login?error=Invalid reset link", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		h.resetPasswordTmpl.Execute(w, AuthPageData{Token: token})
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if !h.authService.ValidatePassword(password) {
		h.resetPasswordTmpl.Execute(w, AuthPageData{Error: "Password must be 8-128 characters", Token: token})
		return
	}
	if password != confirmPassword {
		h.resetPasswordTmpl.Execute(w, AuthPageData{Error: "Passwords do not match", Token: token})
		return
	}

	if err := h.authService.ResetPassword(token, password); err != nil {
		h.resetPasswordTmpl.Execute(w, AuthPageData{Error: err.Error(), Token: token})
		return
	}

	http.Redirect(w, r, "/login?success=Password reset successful! You can now log in.", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		h.authService.Logout(cookie.Value)
	}
	h.authService.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
