package services

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type GmailSMTPService struct {
	username string
	password string
	host     string
	port     string
}

func NewGmailSMTPService() *GmailSMTPService {
	return &GmailSMTPService{
		username: os.Getenv("GMAIL_USERNAME"),
		password: os.Getenv("GMAIL_APP_PASSWORD"),
		host:     "smtp.gmail.com",
		port:     "587",
	}
}

func (s *GmailSMTPService) SendEmail(to, subject, htmlBody string) error {
	if s.username == "" || s.password == "" {
		fmt.Printf("⚠️  Gmail SMTP not configured. Set GMAIL_USERNAME and GMAIL_APP_PASSWORD in .env\n")
		return fmt.Errorf("Gmail SMTP not configured")
	}

	from := s.username
	
	// Build email message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)

	// Authentication
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(sb.String()))
	
	if err != nil {
		fmt.Printf("❌ Failed to send email via Gmail SMTP: %v\n", err)
		return err
	}

	fmt.Printf("✅ Email sent successfully via Gmail SMTP to: %s\n", to)
	return nil
}
