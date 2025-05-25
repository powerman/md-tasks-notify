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

func TestSendEmail(tt *testing.T) {
	t := check.T(tt)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	defaultFrom := fmt.Sprintf("md-tasks-notify@%s", hostname)

	tests := []struct {
		name     string
		setup    func(*testing.T)
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
				Host: "localhost",
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
				Host: "localhost",
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
				Host: "localhost",
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
				Host: "localhost",
				Port: 25,
				From: defaultFrom,
			},
			to:      "to@example.com",
			subject: "Test Subject",
			content: "Hello, World!",
			wantErr: fmt.Errorf("send email: %w", ErrMock),
		},
		{
			name: "from env vars",
			setup: func(tt *testing.T) {
				tt.Helper()
				tt.Setenv("SMTP_HOST", "smtp.example.com")
				tt.Setenv("SMTP_PORT", "2525")
				tt.Setenv("SMTP_USERNAME", "user")
				tt.Setenv("SMTP_PASSWORD", "pass")
				tt.Setenv("SMTP_FROM", "from@example.com")
			},
			to:       "to@example.com",
			subject:  "Test Subject",
			content:  "Hello, World!",
			wantAddr: "smtp.example.com:2525",
			wantAuth: true,
			wantFrom: "from@example.com",
			wantTo:   []string{"to@example.com"},
			wantBody: []string{
				"From: from@example.com",
				"To: to@example.com",
				"Subject: Test Subject",
				"Hello, World!",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			t := check.T(tt)

			// Apply test setup if any
			if test.setup != nil {
				test.setup(tt)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSMTP := NewMockSMTPSender(ctrl)

			// Create Email instance with mock
			email := NewEmail(test.config)
			email.sendMail = mockSMTP.SendMail

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
						t.Equal((auth != nil), test.wantAuth)

						// Verify email body contains expected strings
						body := string(msg)
						for _, want := range test.wantBody {
							t.Contains(body, want)
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
				t.Nil(err)
			} else {
				t.True(errors.Is(err, ErrMock))
			}
		})
	}
}
