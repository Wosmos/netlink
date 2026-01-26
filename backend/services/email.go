package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type EmailService struct {
	apiKey       string
	fromEmail    string
	gmailService *GmailSMTPService
	useGmail     bool
}

func NewEmailService() *EmailService {
	resendKey := os.Getenv("RESEND_API_KEY")
	gmailUser := os.Getenv("GMAIL_USERNAME")

	// Use Gmail SMTP if configured, otherwise use Resend
	useGmail := gmailUser != "" && os.Getenv("GMAIL_APP_PASSWORD") != ""

	service := &EmailService{
		apiKey:       resendKey,
		fromEmail:    os.Getenv("FROM_EMAIL"),
		gmailService: NewGmailSMTPService(),
		useGmail:     useGmail,
	}

	if useGmail {
		fmt.Printf("📧 Email service initialized with Gmail SMTP\n")
	} else if resendKey != "" {
		fmt.Printf("📧 Email service initialized with Resend API\n")
	} else {
		fmt.Printf("⚠️  No email service configured (development mode)\n")
	}

	return service
}

type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *EmailService) Send(to, subject, html string) error {
	fmt.Printf("🔍 Email Service Debug:\n")
	fmt.Printf("  Using: %s\n", func() string {
		if s.useGmail {
			return "Gmail SMTP"
		}
		return "Resend API"
	}())
	fmt.Printf("  From Email: %s\n", s.fromEmail)
	fmt.Printf("  To: %s\n", to)
	fmt.Printf("  Subject: %s\n", subject)

	// Use Gmail SMTP if configured
	if s.useGmail {
		return s.gmailService.SendEmail(to, subject, html)
	}

	// Otherwise use Resend API
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
		fmt.Printf("❌ JSON Marshal error: %v\n", err)
		return err
	}

	fmt.Printf("📤 Request payload: %s\n", string(body))

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("❌ HTTP request creation error: %v\n", err)
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("📤 Sending email via Resend API...\n")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ HTTP request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("📬 Resend API response: %d\n", resp.StatusCode)
	fmt.Printf("📬 Response body: %s\n", string(respBody))

	if resp.StatusCode >= 400 {
		// Check if it's a common Resend restriction error
		if resp.StatusCode == 403 || resp.StatusCode == 422 {
			fmt.Printf("⚠️  This might be a Resend account limitation. Free tier only allows sending to verified email addresses.\n")
			fmt.Printf("⚠️  To send to other emails, upgrade your Resend plan or verify the recipient email in your Resend dashboard.\n")
		}
		return fmt.Errorf("resend API error: %d - %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("✅ Email sent successfully!\n")
	return nil
}

func (s *EmailService) SendVerificationEmail(to, token string) error {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", os.Getenv("APP_URL"), token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
    </style>
</head>
<body style="background-color: #f8fafc; padding: 20px 0;">
    <div style="max-width: 600px; margin: 0 auto; background: white; border-radius: 16px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); overflow: hidden;">
        <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px 30px; text-align: center;">
            <div style="width: 80px; height: 80px; background: rgba(255,255,255,0.2); border-radius: 50%%; margin: 0 auto 20px; display: flex; align-items: center; justify-content: center;">
                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
            </div>
            <h1 style="color: white; font-size: 28px; font-weight: 700; margin-bottom: 8px;">Verify Your Email</h1>
            <p style="color: rgba(255,255,255,0.9); font-size: 16px; font-weight: 400;">Complete your account setup</p>
        </div>
        
        <div style="padding: 40px 30px;">
            <p style="color: #374151; font-size: 16px; line-height: 1.6; margin-bottom: 24px;">Hi there! 👋</p>
            <p style="color: #374151; font-size: 16px; line-height: 1.6; margin-bottom: 32px;">Thanks for signing up! To complete your registration and start using your account, please verify your email address by clicking the button below.</p>
            
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; text-decoration: none; padding: 16px 32px; border-radius: 12px; font-weight: 600; font-size: 16px; box-shadow: 0 4px 14px 0 rgba(102, 126, 234, 0.4);">✨ Verify Email Address</a>
            </div>
            
            <p style="color: #6b7280; font-size: 14px; line-height: 1.5; margin-bottom: 24px;">If the button doesn't work, you can copy and paste this link into your browser:</p>
            <div style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; word-break: break-all;">
                <a href="%s" style="color: #4f46e5; text-decoration: none; font-size: 14px;">%s</a>
            </div>
        </div>
        
        <div style="background: #f9fafb; padding: 24px 30px; text-align: center; border-top: 1px solid #e5e7eb;">
            <p style="color: #6b7280; font-size: 14px; margin-bottom: 8px;">Need help? Contact our support team</p>
            <p style="color: #9ca3af; font-size: 12px;">© 2024 Your App Name. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, verifyURL, verifyURL, verifyURL)

	return s.Send(to, "✨ Verify your email address", html)
}

func (s *EmailService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("APP_URL"), token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
    </style>
</head>
<body style="background-color: #f8fafc; padding: 20px 0;">
    <div style="max-width: 600px; margin: 0 auto; background: white; border-radius: 16px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); overflow: hidden;">
        <div style="background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); padding: 40px 30px; text-align: center;">
            <div style="width: 80px; height: 80px; background: rgba(255,255,255,0.2); border-radius: 50%%; margin: 0 auto 20px; display: flex; align-items: center; justify-content: center;">
                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M15 12C15 13.6569 13.6569 15 12 15C10.3431 15 9 13.6569 9 12C9 10.3431 10.3431 9 12 9C13.6569 9 15 10.3431 15 12Z" stroke="white" stroke-width="2"/>
                </svg>
            </div>
            <h1 style="color: white; font-size: 28px; font-weight: 700; margin-bottom: 8px;">Reset Your Password</h1>
            <p style="color: rgba(255,255,255,0.9); font-size: 16px; font-weight: 400;">Secure your account</p>
        </div>
        
        <div style="padding: 40px 30px;">
            <p style="color: #374151; font-size: 16px; line-height: 1.6; margin-bottom: 24px;">Hi there! 🔐</p>
            <p style="color: #374151; font-size: 16px; line-height: 1.6; margin-bottom: 32px;">We received a request to reset your password. If you made this request, click the button below to create a new password.</p>
            
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); color: white; text-decoration: none; padding: 16px 32px; border-radius: 12px; font-weight: 600; font-size: 16px; box-shadow: 0 4px 14px 0 rgba(245, 87, 108, 0.4);">🔑 Reset Password</a>
            </div>
            
            <p style="color: #6b7280; font-size: 14px; line-height: 1.5; margin-bottom: 24px;">If the button doesn't work, you can copy and paste this link into your browser:</p>
            <div style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; word-break: break-all;">
                <a href="%s" style="color: #4f46e5; text-decoration: none; font-size: 14px;">%s</a>
            </div>
            
            <div style="margin-top: 32px; padding: 20px; background: #fef3c7; border: 1px solid #f59e0b; border-radius: 12px;">
                <p style="color: #92400e; font-size: 14px; line-height: 1.5;">⚠️ This password reset link will expire in <strong>1 hour</strong> for your security.</p>
            </div>
        </div>
        
        <div style="background: #f9fafb; padding: 24px 30px; text-align: center; border-top: 1px solid #e5e7eb;">
            <p style="color: #6b7280; font-size: 14px; margin-bottom: 8px;">Need help? Contact our support team</p>
            <p style="color: #9ca3af; font-size: 12px;">© 2024 Your App Name. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, resetURL, resetURL, resetURL)

	return s.Send(to, "🔑 Reset your password", html)
}

func (s *EmailService) SendOTPEmail(to, otp string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Your Verification Code</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
    </style>
</head>
<body style="background-color: #f8fafc; padding: 20px 0;">
    <div style="max-width: 600px; margin: 0 auto; background: white; border-radius: 16px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); overflow: hidden;">
        <div style="background: linear-gradient(135deg, #4facfe 0%%, #00f2fe 100%%); padding: 40px 30px; text-align: center;">
            <div style="width: 80px; height: 80px; background: rgba(255,255,255,0.2); border-radius: 50%%; margin: 0 auto 20px; display: flex; align-items: center; justify-content: center;">
                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
            </div>
            <h1 style="color: white; font-size: 28px; font-weight: 700; margin-bottom: 8px;">Verification Code</h1>
            <p style="color: rgba(255,255,255,0.9); font-size: 16px; font-weight: 400;">Secure access code</p>
        </div>
        
        <div style="padding: 40px 30px;">
            <p style="color: #374151; font-size: 16px; line-height: 1.6; margin-bottom: 24px;">Hi there! 🔢</p>
            <p style="color: #374151; font-size: 16px; line-height: 1.6; margin-bottom: 32px;">Here's your verification code. Enter this code to complete your verification process:</p>
            
            <div style="text-align: center; margin: 32px 0;">
                <div style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 24px 32px; border-radius: 16px; box-shadow: 0 8px 32px rgba(102, 126, 234, 0.3);">
                    <div style="font-size: 36px; font-weight: 700; letter-spacing: 8px; color: white; font-family: 'Courier New', monospace;">%s</div>
                </div>
            </div>
            
            <div style="margin-top: 32px; padding: 20px; background: #fef3c7; border: 1px solid #f59e0b; border-radius: 12px;">
                <p style="color: #92400e; font-size: 14px; line-height: 1.5;">⏰ This verification code will expire in <strong>10 minutes</strong> for your security.</p>
            </div>
        </div>
        
        <div style="background: #f9fafb; padding: 24px 30px; text-align: center; border-top: 1px solid #e5e7eb;">
            <p style="color: #6b7280; font-size: 14px; margin-bottom: 8px;">Need help? Contact our support team</p>
            <p style="color: #9ca3af; font-size: 12px;">© 2024 Your App Name. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
	`, otp)

	return s.Send(to, "🔢 Your verification code", html)
}
