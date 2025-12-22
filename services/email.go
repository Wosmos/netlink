package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type EmailService struct {
	apiKey    string
	fromEmail string
}

func NewEmailService() *EmailService {
	return &EmailService{
		apiKey:    os.Getenv("RESEND_API_KEY"),
		fromEmail: os.Getenv("FROM_EMAIL"),
	}
}

type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *EmailService) Send(to, subject, html string) error {
	if s.apiKey == "" {
		// Skip if no API key (development mode)
		fmt.Printf("[EMAIL] To: %s, Subject: %s\n", to, subject)
		return nil
	}

	payload := ResendRequest{
		From:    s.fromEmail,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: %d", resp.StatusCode)
	}

	return nil
}

func (s *EmailService) SendVerificationEmail(to, token string) error {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", os.Getenv("APP_URL"), token)

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Verify your email</h2>
			<p>Click the button below to verify your email address:</p>
			<a href="%s" style="display: inline-block; padding: 12px 24px; background: #4f46e5; color: white; text-decoration: none; border-radius: 8px;">
				Verify Email
			</a>
			<p style="margin-top: 20px; color: #666;">
				Or copy this link: %s
			</p>
		</div>
	`, verifyURL, verifyURL)

	return s.Send(to, "Verify your email", html)
}

func (s *EmailService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("APP_URL"), token)

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Reset your password</h2>
			<p>Click the button below to reset your password:</p>
			<a href="%s" style="display: inline-block; padding: 12px 24px; background: #4f46e5; color: white; text-decoration: none; border-radius: 8px;">
				Reset Password
			</a>
			<p style="margin-top: 20px; color: #666;">
				This link expires in 1 hour.
			</p>
		</div>
	`, resetURL)

	return s.Send(to, "Reset your password", html)
}

func (s *EmailService) SendOTPEmail(to, otp string) error {
	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>Your verification code</h2>
			<p>Use this code to verify your identity:</p>
			<div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; padding: 20px; background: #f3f4f6; border-radius: 8px; text-align: center;">
				%s
			</div>
			<p style="margin-top: 20px; color: #666;">
				This code expires in 10 minutes.
			</p>
		</div>
	`, otp)

	return s.Send(to, "Your verification code", html)
}
