package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"testing"
)

// errMock is used to test error handling.
var errMock = errors.New("mock error")

// checkSendMail verifies that SendEmail was called with expected parameters.
func checkSendMail(t *testing.T, tt *struct {
	name       string
	setup      func()
	config     *EmailConfig
	to         string
	subject    string
	content    string
	wantAddr   string
	wantAuth   bool
	wantFrom   string
	wantTo     []string
	wantBody   []string
	wantErrStr string
},
) func(string, smtp.Auth, string, []string, []byte) error {
	t.Helper()
	return func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		if addr != tt.wantAddr {
			t.Errorf("SendEmail() addr = %q, want %q", addr, tt.wantAddr)
		}
		if (auth != nil) != tt.wantAuth {
			t.Errorf("SendEmail() auth = %v, want %v", auth != nil, tt.wantAuth)
		}
		if from != tt.wantFrom {
			t.Errorf("SendEmail() from = %q, want %q", from, tt.wantFrom)
		}
		if !equalSlice(to, tt.wantTo) {
			t.Errorf("SendEmail() to = %q, want %q", to, tt.wantTo)
		}
		for _, want := range tt.wantBody {
			if !strings.Contains(string(msg), want) {
				t.Errorf("SendEmail() body missing %q", want)
			}
		}
		return nil
	}
}

func TestSendEmail(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	defaultFrom := fmt.Sprintf("md-tasks-notify@%s", hostname)

	tests := []struct {
		name       string
		setup      func()
		config     *EmailConfig
		to         string
		subject    string
		content    string
		wantAddr   string
		wantAuth   bool
		wantFrom   string
		wantTo     []string
		wantBody   []string // Strings that should be present in email body
		wantErrStr string
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
			to:         "to@example.com",
			subject:    "Test Subject",
			content:    "Hello, World!",
			wantErrStr: "failed to send email: mock error",
		},
		{
			name: "from env vars",
			setup: func() {
				os.Clearenv()
				os.Setenv("SMTP_HOST", "smtp.example.com")
				os.Setenv("SMTP_PORT", "2525")
				os.Setenv("SMTP_USERNAME", "user")
				os.Setenv("SMTP_PASSWORD", "pass")
				os.Setenv("SMTP_FROM", "from@example.com")
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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Apply test setup if any
			if test.setup != nil {
				test.setup()
			}

			// Create Email instance with mock
			email := NewEmail(test.config)
			if test.wantErrStr == "" {
				email.sendMail = checkSendMail(t, &test)
			} else {
				email.sendMail = func(string, smtp.Auth, string, []string, []byte) error {
					return errMock
				}
			}

			// Run test
			var buf bytes.Buffer
			buf.WriteString(test.content)
			err := email.Send(test.to, test.subject, &buf)

			// Check error
			if test.wantErrStr == "" {
				if err != nil {
					t.Errorf("SendEmail() error = %v, want nil", err)
				}
			} else {
				if err == nil || err.Error() != test.wantErrStr {
					t.Errorf("SendEmail() error = %v, want %q", err, test.wantErrStr)
				}
			}
		})
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
