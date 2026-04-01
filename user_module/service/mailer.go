package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"

	"auto_park/internal/config"
)

type Mailer interface {
	SendWelcome(ctx context.Context, to string, password string, roleName string) error
}

type SMTPMailer struct {
	cfg *config.Config
}

func NewSMTPMailer(cfg *config.Config) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendWelcome(ctx context.Context, to string, password string, roleName string) error {
	if !m.cfg.Email.SendEmail {
		log.Printf("[MAIL] disabled; to=%s role=%s", to, roleName)
		return nil
	}

	host := m.cfg.Email.SMTPHost
	port := m.cfg.Email.SMTPPort
	user := m.cfg.Email.SMTPUser
	pass := m.cfg.Email.SMTPPassword
	fromName := m.cfg.Email.FromName
	fromAddr := m.cfg.Email.FromAddr

	if host == "" || port == 0 || user == "" || pass == "" || fromAddr == "" {
		return errors.New("smtp config missing")
	}

	fromHeader := fromAddr
	if strings.TrimSpace(fromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", fromName, fromAddr)
	}

	subject := "Your Autopark Account"
	bodyJSON, _ := json.MarshalIndent(map[string]any{
		"email":    to,
		"role":     roleName,
		"password": password,
	}, "", "  ")

	headers := map[string]string{
		"From":         fromHeader,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=utf-8",
	}

	var sb strings.Builder
	for k, v := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	sb.WriteString("\r\n")
	sb.WriteString("Welcome to Autopark!\nBelow are your credentials:\n\n")
	sb.WriteString(string(bodyJSON))

	msg := []byte(sb.String())

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	tlsCfg := &tls.Config{ServerName: host}
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	} else {
		return errors.New("server does not support STARTTLS")
	}

	if ok, _ := c.Extension("AUTH"); ok {
		if err := c.Auth(smtp.PlainAuth("", user, pass, host)); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := c.Mail(fromAddr); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, _ = wc.Write(msg)
	if err := wc.Close(); err != nil {
		return fmt.Errorf("data close: %w", err)
	}

	_ = c.Quit()
	return nil
}
