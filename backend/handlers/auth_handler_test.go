package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"netlink/models"
)

// APIResponse is the standard JSON response shape from the API.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
}

// mockAuthService implements AuthServiceInterface for testing
type mockAuthService struct {
	loginFn             func(email, password string) (*models.Session, error)
	registerFn          func(email, password string, name, phone *string) error
	verifyEmailFn       func(token string) error
	forgotPasswordFn    func(email string) error
	resetPasswordFn     func(token, newPassword string) error
	getUserFromReqFn    func(r *http.Request) (*models.User, error)
	getUserFromSessFn   func(session *models.Session) (*models.User, error)
	setSessionCookieFn  func(w http.ResponseWriter, session *models.Session)
	clearSessionCookFn  func(w http.ResponseWriter)
	validateEmailFn     func(email string) bool
	validatePasswordFn  func(password string) bool
	testEmailFn         func(to string) error
	logoutFn            func(sessionID string) error
}

func (m *mockAuthService) Login(email, password string) (*models.Session, error) {
	if m.loginFn != nil {
		return m.loginFn(email, password)
	}
	return nil, nil
}

func (m *mockAuthService) Register(email, password string, name, phone *string) error {
	if m.registerFn != nil {
		return m.registerFn(email, password, name, phone)
	}
	return nil
}

func (m *mockAuthService) VerifyEmail(token string) error {
	if m.verifyEmailFn != nil {
		return m.verifyEmailFn(token)
	}
	return nil
}

func (m *mockAuthService) ForgotPassword(email string) error {
	if m.forgotPasswordFn != nil {
		return m.forgotPasswordFn(email)
	}
	return nil
}

func (m *mockAuthService) ResetPassword(token, newPassword string) error {
	if m.resetPasswordFn != nil {
		return m.resetPasswordFn(token, newPassword)
	}
	return nil
}

func (m *mockAuthService) GetUserFromRequest(r *http.Request) (*models.User, error) {
	if m.getUserFromReqFn != nil {
		return m.getUserFromReqFn(r)
	}
	return nil, nil
}

func (m *mockAuthService) GetUserFromSession(session *models.Session) (*models.User, error) {
	if m.getUserFromSessFn != nil {
		return m.getUserFromSessFn(session)
	}
	return nil, nil
}

func (m *mockAuthService) SetSessionCookie(w http.ResponseWriter, session *models.Session) {
	if m.setSessionCookieFn != nil {
		m.setSessionCookieFn(w, session)
	}
}

func (m *mockAuthService) ClearSessionCookie(w http.ResponseWriter) {
	if m.clearSessionCookFn != nil {
		m.clearSessionCookFn(w)
	}
}

func (m *mockAuthService) ValidateEmail(email string) bool {
	if m.validateEmailFn != nil {
		return m.validateEmailFn(email)
	}
	return true
}

func (m *mockAuthService) ValidatePassword(password string) bool {
	if m.validatePasswordFn != nil {
		return m.validatePasswordFn(password)
	}
	return true
}

func (m *mockAuthService) TestEmailDelivery(to string) error {
	if m.testEmailFn != nil {
		return m.testEmailFn(to)
	}
	return nil
}

func (m *mockAuthService) Logout(sessionID string) error {
	if m.logoutFn != nil {
		return m.logoutFn(sessionID)
	}
	return nil
}

// helper to create handler with mock
func newTestAuthHandler(mock *mockAuthService) *AuthHandler {
	return NewAPIAuthHandler(mock)
}

// helper to decode API response
func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) APIResponse {
	t.Helper()
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	return resp
}

// =========== APILogin Tests ===========

func TestAPILogin_Success(t *testing.T) {
	session := &models.Session{
		ID:        "test-session-id",
		UserID:    1,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	user := &models.User{ID: 1, Email: "user@test.com", Name: "Test User"}

	mock := &mockAuthService{
		loginFn: func(email, password string) (*models.Session, error) {
			if email != "user@test.com" || password != "password123" {
				t.Errorf("Login called with wrong args: email=%q, password=%q", email, password)
			}
			return session, nil
		},
		setSessionCookieFn: func(w http.ResponseWriter, s *models.Session) {
			if s.ID != session.ID {
				t.Errorf("SetSessionCookie called with wrong session ID: %q", s.ID)
			}
		},
		getUserFromSessFn: func(s *models.Session) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"user@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APILogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if raw["success"] != true {
		t.Error("Expected success=true")
	}
	if token, _ := raw["token"].(string); token != "test-session-id" {
		t.Errorf("Expected token=%q, got %q", "test-session-id", token)
	}
}

func TestAPILogin_InvalidCredentials(t *testing.T) {
	mock := &mockAuthService{
		loginFn: func(email, password string) (*models.Session, error) {
			return nil, nil // nil session = invalid creds
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"wrong@test.com","password":"badpass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APILogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if resp.Success {
		t.Error("Expected success=false")
	}
	if resp.Error != "Invalid email or password" {
		t.Errorf("Expected error=%q, got %q", "Invalid email or password", resp.Error)
	}
}

func TestAPILogin_UnverifiedEmail(t *testing.T) {
	mock := &mockAuthService{
		loginFn: func(email, password string) (*models.Session, error) {
			return nil, errors.New("please verify your email first")
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"user@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APILogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if resp.Error != "please verify your email first" {
		t.Errorf("Expected verify error, got %q", resp.Error)
	}
}

func TestAPILogin_InvalidJSON(t *testing.T) {
	mock := &mockAuthService{}
	h := newTestAuthHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.APILogin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPILogin_MethodNotAllowed(t *testing.T) {
	mock := &mockAuthService{}
	h := newTestAuthHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	w := httptest.NewRecorder()

	h.APILogin(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// =========== APIRegister Tests ===========

func TestAPIRegister_Success(t *testing.T) {
	var capturedEmail, capturedPassword string
	var capturedName *string

	mock := &mockAuthService{
		validateEmailFn:    func(email string) bool { return true },
		validatePasswordFn: func(password string) bool { return true },
		registerFn: func(email, password string, name, phone *string) error {
			capturedEmail = email
			capturedPassword = password
			capturedName = name
			return nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"new@test.com","password":"strongpass1","name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIRegister(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("Expected success=true")
	}
	if resp.Message != "Verification email sent" {
		t.Errorf("Expected message=%q, got %q", "Verification email sent", resp.Message)
	}
	if capturedEmail != "new@test.com" {
		t.Errorf("Expected email=%q, got %q", "new@test.com", capturedEmail)
	}
	if capturedPassword != "strongpass1" {
		t.Errorf("Expected password=%q, got %q", "strongpass1", capturedPassword)
	}
	if capturedName == nil || *capturedName != "Alice" {
		t.Error("Expected name=Alice")
	}
}

func TestAPIRegister_InvalidEmail(t *testing.T) {
	mock := &mockAuthService{
		validateEmailFn: func(email string) bool { return false },
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"bademail","password":"strongpass1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if resp.Error != "Invalid email" {
		t.Errorf("Expected error=%q, got %q", "Invalid email", resp.Error)
	}
}

func TestAPIRegister_WeakPassword(t *testing.T) {
	mock := &mockAuthService{
		validateEmailFn:    func(email string) bool { return true },
		validatePasswordFn: func(password string) bool { return false },
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"user@test.com","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if resp.Error != "Password must be 8-128 characters" {
		t.Errorf("Expected password error, got %q", resp.Error)
	}
}

func TestAPIRegister_DuplicateEmail(t *testing.T) {
	mock := &mockAuthService{
		validateEmailFn:    func(email string) bool { return true },
		validatePasswordFn: func(password string) bool { return true },
		registerFn: func(email, password string, name, phone *string) error {
			return errors.New("UNIQUE constraint violation")
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"existing@test.com","password":"strongpass1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if resp.Error != "Email already registered" {
		t.Errorf("Expected error=%q, got %q", "Email already registered", resp.Error)
	}
}

func TestAPIRegister_MethodNotAllowed(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodGet, "/api/auth/register", nil)
	w := httptest.NewRecorder()

	h.APIRegister(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// =========== Me Tests ===========

func TestMe_Authenticated(t *testing.T) {
	user := &models.User{ID: 42, Email: "me@test.com", Name: "Test User"}
	mock := &mockAuthService{
		getUserFromReqFn: func(r *http.Request) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("Expected success=true")
	}
	if resp.Data == nil {
		t.Fatal("Expected user data")
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	mock := &mockAuthService{
		getUserFromReqFn: func(r *http.Request) (*models.User, error) {
			return nil, nil
		},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if resp.Success {
		t.Error("Expected success=false")
	}
}

// =========== APILogout Tests ===========

func TestAPILogout_Success(t *testing.T) {
	var loggedOutSessionID string
	mock := &mockAuthService{
		logoutFn: func(sessionID string) error {
			loggedOutSessionID = sessionID
			return nil
		},
		clearSessionCookFn: func(w http.ResponseWriter) {},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-abc"})
	w := httptest.NewRecorder()

	h.APILogout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if loggedOutSessionID != "sess-abc" {
		t.Errorf("Expected logout session=%q, got %q", "sess-abc", loggedOutSessionID)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestAPILogout_NoCookie(t *testing.T) {
	mock := &mockAuthService{
		clearSessionCookFn: func(w http.ResponseWriter) {},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()

	h.APILogout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// =========== APIVerify Tests ===========

func TestAPIVerify_Success(t *testing.T) {
	mock := &mockAuthService{
		verifyEmailFn: func(token string) error {
			if token != "verify-token-123" {
				t.Errorf("Expected token=%q, got %q", "verify-token-123", token)
			}
			return nil
		},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify?token=verify-token-123", nil)
	w := httptest.NewRecorder()

	h.APIVerify(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestAPIVerify_MissingToken(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify", nil)
	w := httptest.NewRecorder()

	h.APIVerify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPIVerify_InvalidToken(t *testing.T) {
	mock := &mockAuthService{
		verifyEmailFn: func(token string) error {
			return errors.New("invalid verification token")
		},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/verify?token=bad-token", nil)
	w := httptest.NewRecorder()

	h.APIVerify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPIVerify_MethodNotAllowed(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/verify", nil)
	w := httptest.NewRecorder()

	h.APIVerify(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// =========== APIForgotPassword Tests ===========

func TestAPIForgotPassword_Success(t *testing.T) {
	mock := &mockAuthService{
		validateEmailFn: func(email string) bool { return true },
		forgotPasswordFn: func(email string) error {
			return nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"user@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIForgotPassword(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestAPIForgotPassword_InvalidEmail(t *testing.T) {
	mock := &mockAuthService{
		validateEmailFn: func(email string) bool { return false },
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"notanemail"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIForgotPassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPIForgotPassword_ServiceError(t *testing.T) {
	mock := &mockAuthService{
		validateEmailFn: func(email string) bool { return true },
		forgotPasswordFn: func(email string) error {
			return errors.New("db error")
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"user@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIForgotPassword(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// =========== APIResetPassword Tests ===========

func TestAPIResetPassword_Success(t *testing.T) {
	mock := &mockAuthService{
		validatePasswordFn: func(password string) bool { return true },
		resetPasswordFn: func(token, newPassword string) error {
			if token != "reset-token" {
				t.Errorf("Expected token=%q, got %q", "reset-token", token)
			}
			return nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"token":"reset-token","password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIResetPassword(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	resp := decodeResponse(t, w)
	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestAPIResetPassword_MissingToken(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{})
	body := `{"token":"","password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIResetPassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPIResetPassword_WeakPassword(t *testing.T) {
	mock := &mockAuthService{
		validatePasswordFn: func(password string) bool { return false },
	}

	h := newTestAuthHandler(mock)
	body := `{"token":"reset-token","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIResetPassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAPIResetPassword_ExpiredToken(t *testing.T) {
	mock := &mockAuthService{
		validatePasswordFn: func(password string) bool { return true },
		resetPasswordFn: func(token, newPassword string) error {
			return errors.New("invalid or expired reset token")
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"token":"expired-token","password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APIResetPassword(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// =========== TestEmail Tests ===========

func TestTestEmail_Success(t *testing.T) {
	mock := &mockAuthService{
		testEmailFn: func(to string) error {
			if to != "test@example.com" {
				t.Errorf("Expected to=%q, got %q", "test@example.com", to)
			}
			return nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/test-email", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.TestEmail(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestTestEmail_DefaultAddress(t *testing.T) {
	var sentTo string
	mock := &mockAuthService{
		testEmailFn: func(to string) error {
			sentTo = to
			return nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/test-email", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.TestEmail(w, req)

	if sentTo != "delivered@resend.dev" {
		t.Errorf("Expected default to=%q, got %q", "delivered@resend.dev", sentTo)
	}
}

func TestTestEmail_ServiceError(t *testing.T) {
	mock := &mockAuthService{
		testEmailFn: func(to string) error {
			return errors.New("smtp failed")
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/test-email", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.TestEmail(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestTestEmail_MethodNotAllowed(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodGet, "/api/test-email", nil)
	w := httptest.NewRecorder()

	h.TestEmail(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// =========== Response Format Tests ===========

func TestAPIResponse_ContentType(t *testing.T) {
	mock := &mockAuthService{
		getUserFromReqFn: func(r *http.Request) (*models.User, error) {
			return nil, nil
		},
	}

	h := newTestAuthHandler(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type=%q, got %q", "application/json", contentType)
	}
}

func TestAPILogin_EmailNormalization(t *testing.T) {
	var capturedEmail string
	mock := &mockAuthService{
		loginFn: func(email, password string) (*models.Session, error) {
			capturedEmail = email
			return nil, nil
		},
	}

	h := newTestAuthHandler(mock)
	body := `{"email":"  USER@TEST.COM  ","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.APILogin(w, req)

	if capturedEmail != "user@test.com" {
		t.Errorf("Expected normalized email=%q, got %q", "user@test.com", capturedEmail)
	}
}
