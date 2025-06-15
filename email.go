package main

import (
	"fmt"
	"io"
	"log"
	"net/smtp"
	"os"
	"strconv"
)

const (
	defaultFrom        = "md-tasks-notify"
	smtpPort           = 25
	smtpSubmissionPort = 587
)

// EmailConfig holds configuration for sending emails.
type EmailConfig struct {
	Host     string
	Port     int
	Username string // May be empty when auth not needed.
	Password string // May be empty when auth not needed.
	From     string
}

// NewEmailConfigFromEnv returns email configuration from environment variables.
// If environment variables are not set, returns config for localhost:25 without auth.
func NewEmailConfigFromEnv() *EmailConfig {
	cfg := &EmailConfig{
		From:     os.Getenv("SMTP_FROM"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		Host:     os.Getenv("SMTP_HOST"),
		Port:     smtpPort,
	}
	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		portStr = "0"
	}
	port, portErr := strconv.Atoi(portStr)

	if cfg.From == "" {
		cfg.From = defaultFrom
	}
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Username != "" || cfg.Password != "" {
		cfg.Port = smtpSubmissionPort // Use submission port when auth required.
	}
	switch {
	case portErr != nil:
		log.Println("Warning: Ignoring SMTP port:", portErr)
	case port > 0 && port < 65536:
		cfg.Port = port
	case port != 0:
		log.Printf("Warning: Ignoring invalid SMTP port %d", port)
	}

	return cfg
}

// Email provides email sending functionality.
type Email struct {
	cfg      *EmailConfig
	sendMail func(string, smtp.Auth, string, []string, []byte) error
}

// NewEmail creates a new Email instance.
func NewEmail(cfg *EmailConfig) *Email {
	if cfg == nil {
		cfg = NewEmailConfigFromEnv()
	}
	if cfg.From == "" || cfg.Host == "" || cfg.Port == 0 {
		panic(fmt.Sprintf("Invalid EmailConfig: %+v", cfg))
	}
	return &Email{
		cfg:      cfg,
		sendMail: smtp.SendMail,
	}
}

// Send sends email with given content to specified recipient.
func (e *Email) Send(to, subject string, content io.Reader) error {
	// Read content into buffer
	body, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("read email content: %w", err)
	}

	// Compose email message
	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", to, e.cfg.From, subject, body)

	// Connect to SMTP server
	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)

	if err := e.sendMail(addr, auth, e.cfg.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
