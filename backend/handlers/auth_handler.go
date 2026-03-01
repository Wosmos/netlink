package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"netlink/auth"
	"netlink/middleware"
)

type AuthHandler struct {
	authService        AuthServiceInterface
	loginTmpl          *template.Template
	registerTmpl       *template.Template
	forgotPasswordTmpl *template.Template
	resetPasswordTmpl  *template.Template
}

func NewAuthHandler(authService AuthServiceInterface) *AuthHandler {
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

// NewAPIAuthHandler creates an AuthHandler without templates (for API-only use and testing).
func NewAPIAuthHandler(authService AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
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
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))

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

	// Prepare optional fields
	var namePtr, phonePtr *string
	if name != "" {
		namePtr = &name
	}
	if phone != "" {
		phonePtr = &phone
	}

	if err := h.authService.Register(email, password, namePtr, phonePtr); err != nil {
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

// ============ API Methods ============

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

// GET /api/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.GetUserFromRequest(r)
	if err != nil || user == nil {
		middleware.JSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	middleware.JSONSuccess(w, user)
}

// POST /api/auth/login
func (h *AuthHandler) APILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONMethodNotAllowed(w)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	session, err := h.authService.Login(email, req.Password)
	if err != nil {
		middleware.JSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if session == nil {
		middleware.JSONError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	h.authService.SetSessionCookie(w, session)

	user, _ := h.authService.GetUserFromSession(session)
	middleware.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    user,
		"token":   session.ID,
	})
}

// POST /api/auth/register
func (h *AuthHandler) APIRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONMethodNotAllowed(w)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !h.authService.ValidateEmail(email) {
		middleware.JSONError(w, "Invalid email", http.StatusBadRequest)
		return
	}
	if !h.authService.ValidatePassword(req.Password) {
		middleware.JSONError(w, "Password must be 8-128 characters", http.StatusBadRequest)
		return
	}

	// Prepare optional fields
	var namePtr, phonePtr *string
	if req.Name != "" {
		name := strings.TrimSpace(req.Name)
		namePtr = &name
	}
	if req.Phone != "" {
		phone := strings.TrimSpace(req.Phone)
		phonePtr = &phone
	}

	if err := h.authService.Register(email, req.Password, namePtr, phonePtr); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			middleware.JSONError(w, "Email already registered", http.StatusBadRequest)
		} else {
			middleware.JSONError(w, "Registration failed", http.StatusBadRequest)
		}
		return
	}

	middleware.JSONSuccessMessage(w, "Verification email sent")
}

// POST /api/auth/logout
func (h *AuthHandler) APILogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
		h.authService.Logout(cookie.Value)
	}
	h.authService.ClearSessionCookie(w)

	middleware.JSONSuccessMessage(w, "Logged out")
}

// POST /api/auth/forgot-password
func (h *AuthHandler) APIForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONMethodNotAllowed(w)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !h.authService.ValidateEmail(email) {
		middleware.JSONError(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if err := h.authService.ForgotPassword(email); err != nil {
		log.Printf("Forgot password error: %v", err)
		middleware.JSONError(w, "An error occurred", http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccessMessage(w, "If that email is registered, you will receive a reset link")
}

// POST /api/auth/reset-password
func (h *AuthHandler) APIResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONMethodNotAllowed(w)
		return
	}

	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Token == "" {
		middleware.JSONError(w, "Token required", http.StatusBadRequest)
		return
	}

	if !h.authService.ValidatePassword(req.Password) {
		middleware.JSONError(w, "Password must be 8-128 characters", http.StatusBadRequest)
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.Password); err != nil {
		middleware.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	middleware.JSONSuccessMessage(w, "Password reset successful")
}

// GET /api/auth/verify - API version of email verification
func (h *AuthHandler) APIVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.JSONMethodNotAllowed(w)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		middleware.JSONError(w, "Invalid verification link", http.StatusBadRequest)
		return
	}

	if err := h.authService.VerifyEmail(token); err != nil {
		middleware.JSONError(w, "Verification failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	middleware.JSONSuccessMessage(w, "Email verified successfully! You can now log in.")
}

// POST /api/test-email - Test email functionality (dev-only)
func (h *AuthHandler) TestEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.JSONMethodNotAllowed(w)
		return
	}

	// Guard: only allow in development
	if os.Getenv("ENV") == "production" || os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		middleware.JSONError(w, "Not available in production", http.StatusForbidden)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Test with Resend test address first
	testEmail := "delivered@resend.dev"
	if req.Email != "" {
		testEmail = req.Email
	}

	if err := h.authService.TestEmailDelivery(testEmail); err != nil {
		log.Printf("Test email error: %v", err)
		middleware.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	middleware.JSONSuccessMessage(w, fmt.Sprintf("Test email sent to %s", testEmail))
}
