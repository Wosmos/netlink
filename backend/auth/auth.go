package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go-to-do/models"
	"go-to-do/repository"
	"go-to-do/services"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "session_id"
	SessionDuration   = 24 * time.Hour * 7
	BcryptCost        = 8
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)

type AuthService struct {
	userRepo     *repository.UserRepository
	sessionRepo  *repository.SessionRepository
	emailService *services.EmailService
}

func NewAuthService(userRepo *repository.UserRepository, sessionRepo *repository.SessionRepository) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		emailService: services.NewEmailService(),
	}
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
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (s *AuthService) ValidateEmail(email string) bool {
	return emailRegex.MatchString(email) && len(email) <= 254
}

func (s *AuthService) ValidatePassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

func (s *AuthService) Register(email, password string, name, phone *string) error {
	hash, err := s.HashPassword(password)
	if err != nil {
		return err
	}
	token := s.GenerateToken()
	_, err = s.userRepo.Create(email, hash, token, name, phone)
	if err == nil {
		// Send verification email using Resend
		if emailErr := s.emailService.SendVerificationEmail(email, token); emailErr != nil {
			fmt.Printf("Failed to send verification email: %v\n", emailErr)
			// Check if it's a Resend restriction error
			if strings.Contains(emailErr.Error(), "403") || strings.Contains(emailErr.Error(), "422") {
				fmt.Printf("⚠️  Email sending restricted. This is likely because:\n")
				fmt.Printf("   - You're on Resend's free tier which only allows sending to verified emails\n")
				fmt.Printf("   - The recipient email (%s) is not verified in your Resend account\n", email)
				fmt.Printf("   - To send to any email, upgrade your Resend plan\n")
			}
			// Still print to console as fallback
			fmt.Printf("[FALLBACK EMAIL] To: %s, Verify: http://localhost:3000/verify?token=%s\n", email, token)
		} else {
			fmt.Printf("✅ Verification email sent successfully to: %s\n", email)
		}
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
	if err != nil || user == nil {
		return nil
	}
	token := s.GenerateToken()
	expires := time.Now().Add(1 * time.Hour)
	err = s.userRepo.SetResetToken(user.ID, &token, &expires)
	if err == nil {
		// Send password reset email using Resend
		if emailErr := s.emailService.SendPasswordResetEmail(email, token); emailErr != nil {
			fmt.Printf("Failed to send password reset email: %v\n", emailErr)
			// Check if it's a Resend restriction error
			if strings.Contains(emailErr.Error(), "403") || strings.Contains(emailErr.Error(), "422") {
				fmt.Printf("⚠️  Email sending restricted. This is likely because:\n")
				fmt.Printf("   - You're on Resend's free tier which only allows sending to verified emails\n")
				fmt.Printf("   - The recipient email (%s) is not verified in your Resend account\n", email)
				fmt.Printf("   - To send to any email, upgrade your Resend plan\n")
			}
			// Still print to console as fallback
			fmt.Printf("[FALLBACK EMAIL] To: %s, Reset: http://localhost:3000/reset-password?token=%s\n", email, token)
		} else {
			fmt.Printf("✅ Password reset email sent successfully to: %s\n", email)
		}
	}
	return err
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
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
		return nil, nil
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
	var sessionID string

	// 1. Try to get from cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		sessionID = cookie.Value
	}

	// 2. Try to get from Authorization header (for mobile apps)
	if sessionID == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			sessionID = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 3. Try to get from query parameter (for WebSockets)
	if sessionID == "" {
		sessionID = r.URL.Query().Get("token")
	}

	if sessionID == "" {
		return nil, nil
	}

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil || session == nil {
		return nil, err
	}
	return s.userRepo.GetByID(session.UserID)
}

func (s *AuthService) GetUserFromSession(session *models.Session) (*models.User, error) {
	if session == nil {
		return nil, nil
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
		Secure:   false,
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

func (s *AuthService) TestEmailDelivery(to string) error {
	return s.emailService.Send(to, "🧪 Test Email from Your App", `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Test Email</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h2 style="color: #4f46e5;">✅ Email Service Test</h2>
    <p>This is a test email to verify that your Resend integration is working correctly.</p>
    <p><strong>Timestamp:</strong> `+time.Now().Format("2006-01-02 15:04:05")+`</p>
    <p>If you received this email, your email service is configured properly!</p>
</body>
</html>
	`)
}
