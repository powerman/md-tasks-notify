package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"testing"

	"github.com/powerman/check"
	"go.uber.org/mock/gomock"
)

// ErrMock is used to test error handling.
var ErrMock = errors.New("mock error")

func TestSendEmail(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = defaultHost
	}
	defaultFrom := fmt.Sprintf("md-tasks-notify@%s", hostname)

	tests := []struct {
		name     string
		config   *EmailConfig
		to       string
		subject  string
		content  string
		wantAddr string
		wantAuth bool
		wantFrom string
		wantTo   []string
		wantBody []string
		wantErr  error
	}{
		{
			name: "local without auth",
			config: &EmailConfig{
				Host: defaultHost,
				Port: 25,
				From: defaultFrom,
			},
			to:       "to@example.com",
			subject:  "Test Subject",
			content:  "Hello, World!",
			wantAddr: "localhost:25",
			wantAuth: false,
			wantFrom: defaultFrom,
			wantTo:   []string{"to@example.com"},
			wantBody: []string{
				"To: to@example.com",
				"Subject: Test Subject",
				"Hello, World!",
			},
		},
		{
			name: "custom from address",
			config: &EmailConfig{
				Host: defaultHost,
				Port: 25,
				From: "from@example.com",
			},
			to:       "to@example.com",
			subject:  "Test Subject",
			content:  "Hello, World!",
			wantAddr: "localhost:25",
			wantAuth: false,
			wantFrom: "from@example.com",
			wantTo:   []string{"to@example.com"},
			wantBody: []string{
				"From: from@example.com",
				"To: to@example.com",
				"Subject: Test Subject",
				"Hello, World!",
			},
		},
		{
			name: "with auth",
			config: &EmailConfig{
				Host:     "localhost",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "user",
			},
			to:       "to@example.com",
			subject:  "Test Subject",
			content:  "Hello, World!",
			wantAddr: "localhost:587",
			wantAuth: true,
			wantFrom: "user",
			wantTo:   []string{"to@example.com"},
			wantBody: []string{
				"From: user",
				"To: to@example.com",
				"Subject: Test Subject",
				"Hello, World!",
			},
		},
		{
			name: "custom port",
			config: &EmailConfig{
				Host: defaultHost,
				Port: 2525,
				From: defaultFrom,
			},
			to:       "to@example.com",
			subject:  "Test Subject",
			content:  "Hello, World!",
			wantAddr: "localhost:2525",
			wantAuth: false,
			wantFrom: defaultFrom,
			wantTo:   []string{"to@example.com"},
			wantBody: []string{
				"To: to@example.com",
				"Subject: Test Subject",
				"Hello, World!",
			},
		},
		{
			name: "send error",
			config: &EmailConfig{
				Host: defaultHost,
				Port: 25,
				From: defaultFrom,
			},
			to:      "to@example.com",
			subject: "Test Subject",
			content: "Hello, World!",
			wantErr: fmt.Errorf("send email: %w", ErrMock),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := check.Must(t)

			ctrl := gomock.NewController(c)
			t.Cleanup(ctrl.Finish)

			mockSMTP := NewMockSMTPSender(ctrl)

			// Create Email instance with mock
			if test.config == nil {
				test.config = NewEmailConfigFromEnv()
			}
			test.config.SendMail = mockSMTP.SendMail
			email := NewEmail(test.config)

			if test.wantErr == nil {
				mockSMTP.EXPECT().
					SendMail(
						test.wantAddr,
						gomock.Any(),
						test.wantFrom,
						test.wantTo,
						gomock.Any(),
					).
					DoAndReturn(func(_ string, auth smtp.Auth, _ string, _ []string, msg []byte) error {
						// Verify auth
						c.Equal((auth != nil), test.wantAuth)

						// Verify email body contains expected strings
						body := string(msg)
						for _, want := range test.wantBody {
							c.Contains(body, want)
						}
						return nil
					})
			} else {
				mockSMTP.EXPECT().
					SendMail(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(ErrMock)
			}

			// Run test
			var buf bytes.Buffer
			buf.WriteString(test.content)
			err := email.Send(test.to, test.subject, &buf)

			// Check error
			if test.wantErr == nil {
				c.Nil(err)
			} else {
				c.True(errors.Is(err, ErrMock))
			}
		})
	}
}

func TestSendEmailFromEnvVars(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM", "from@example.com")

	c := check.Must(t)

	ctrl := gomock.NewController(c)
	t.Cleanup(ctrl.Finish)

	mockSMTP := NewMockSMTPSender(ctrl)

	config := NewEmailConfigFromEnv()
	config.SendMail = mockSMTP.SendMail
	email := NewEmail(config)

	mockSMTP.EXPECT().
		SendMail(
			"smtp.example.com:2525",
			gomock.Any(),
			"from@example.com",
			[]string{"to@example.com"},
			gomock.Any(),
		).
		DoAndReturn(func(_ string, auth smtp.Auth, _ string, _ []string, msg []byte) error {
			c.Equal((auth != nil), true)

			body := string(msg)
			c.Contains(body, "From: from@example.com")
			c.Contains(body, "To: to@example.com")
			c.Contains(body, "Subject: Test Subject")
			c.Contains(body, "Hello, World!")
			return nil
		})

	var buf bytes.Buffer
	buf.WriteString("Hello, World!")
	err := email.Send("to@example.com", "Test Subject", &buf)
	c.Nil(err)
}
