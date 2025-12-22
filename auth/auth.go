package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"go-to-do/models"
	"go-to-do/repository"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "session_id"
	SessionDuration   = 24 * time.Hour * 7 // 7 days
	BcryptCost        = 8                  // Match seed file cost (12 is too slow for 10K concurrent)
)

var emailRegex = regexp.MustCompile(
	`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`,
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository) *AuthService {
	return &AuthService{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (s *AuthService) GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) ValidateEmail(email string) bool {
	return emailRegex.MatchString(email) && len(email) <= 254
}

func (s *AuthService) ValidatePassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

func (s *AuthService) Register(email, password string) error {
	hash, err := s.HashPassword(password)
	if err != nil {
		return err
	}
	token := s.GenerateToken()
	_, err = s.userRepo.Create(email, hash, token)
	if err == nil {
		// Mock sending email
		fmt.Printf("📧 [MOCK EMAIL] To: %s, Body: Your verification link is http://localhost:8080/verify?token=%s\n", email, token)
	}
	return err
}

func (s *AuthService) VerifyEmail(token string) error {
	user, err := s.userRepo.GetByVerificationToken(token)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("invalid verification token")
	}
	return s.userRepo.UpdateVerificationStatus(user.ID, true)
}

func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		// Don't reveal if email exists, but we can't send if it doesn't
		return nil
	}

	token := s.GenerateToken()
	expires := time.Now().Add(1 * time.Hour)
	err = s.userRepo.SetResetToken(user.ID, &token, &expires)
	if err == nil {
		// Mock sending email
		fmt.Printf("📧 [MOCK EMAIL] To: %s, Body: Your password reset link is http://localhost:8080/reset-password?token=%s\n", email, token)
	}
	return err
}

func (s *AuthService) ResetPassword(token string, newPassword string) error {
	user, err := s.userRepo.GetByResetToken(token)
	if err != nil {
		return err
	}
	if user == nil || user.ResetTokenExpires == nil || user.ResetTokenExpires.Before(time.Now()) {
		return errors.New("invalid or expired reset token")
	}

	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(user.ID, hash)
}

func (s *AuthService) Login(email, password string) (*models.Session, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil || !s.CheckPassword(password, user.PasswordHash) {
		return nil, nil // Invalid credentials
	}
	if !user.IsVerified {
		return nil, errors.New("please verify your email first")
	}
	return s.sessionRepo.Create(user.ID, SessionDuration)
}

func (s *AuthService) Logout(sessionID string) error {
	return s.sessionRepo.Delete(sessionID)
}

func (s *AuthService) GetUserFromRequest(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, nil
	}
	session, err := s.sessionRepo.GetByID(cookie.Value)
	if err != nil || session == nil {
		return nil, err
	}
	return s.userRepo.GetByID(session.UserID)
}

func (s *AuthService) SetSessionCookie(w http.ResponseWriter, session *models.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *AuthService) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
