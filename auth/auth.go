package auth

import (
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
	_, err = s.userRepo.Create(email, hash)
	return err
}

func (s *AuthService) Login(email, password string) (*models.Session, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil || !s.CheckPassword(password, user.PasswordHash) {
		return nil, nil // Invalid credentials
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
