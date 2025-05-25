//go:generate -command MOCKGEN sh -c "$(git rev-parse --show-toplevel)/.buildcache/bin/$DOLLAR{DOLLAR}0 \"$DOLLAR{DOLLAR}@\"" mockgen
//go:generate MOCKGEN -package=$GOPACKAGE -source=$GOFILE -destination=mock.$GOFILE

package main

import "net/smtp"

// SMTPSender defines the interface for sending emails, primarily used for testing.
//
//nolint:iface // SMTPSender is used in tests via mock
type SMTPSender interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// SMTPFunc implements SMTPSender interface.
type SMTPFunc func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

// SendMail implements the SMTPSender interface by calling the underlying function.
func (f SMTPFunc) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return f(addr, a, from, to, msg)
}
