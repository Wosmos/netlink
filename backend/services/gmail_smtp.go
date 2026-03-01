package services

import (
	"fmt"
	"log"
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
		return fmt.Errorf("Gmail SMTP not configured: set GMAIL_USERNAME and GMAIL_APP_PASSWORD")
	}

	from := s.username

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(sb.String())); err != nil {
		return fmt.Errorf("gmail SMTP send to %s: %w", to, err)
	}

	log.Printf("Email sent via Gmail SMTP to %s", to)
	return nil
}
