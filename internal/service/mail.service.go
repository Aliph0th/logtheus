package service

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"logtheus/internal/config"
	"path/filepath"
	"strings"

	"github.com/wneessen/go-mail"
)

type MailService struct {
	client     *mail.Client
	fromHeader string
}

func NewMailService(cfg *config.AppConfig) *MailService {
	client, err := mail.NewClient(
		cfg.SMTP.Host,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(cfg.SMTP.Login),
		mail.WithPassword(cfg.SMTP.Password),
		mail.WithPort(cfg.SMTP.Port),
	)
	if err != nil {
		panic(err)
	}
	return &MailService{client: client, fromHeader: cfg.SMTP.From}
}

func (s *MailService) SendVerifyEmail(to, username, domain, code string) error {
	body, err := s.renderVerifyEmailTemplate(username, domain, code)
	if err != nil {
		return err
	}
	return s.sendMail(to, "Logtheus email verification", body)
}

func (s *MailService) sendMail(to, subject, body string) error {
	message := mail.NewMsg()
	message.From(s.fromHeader)
	message.To(to)
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, body)
	slog.Info("Sending email", "to", to, "subject", subject)
	err := s.client.DialAndSend(message)
	if err != nil {
		return err
	}
	return nil
}

func (s *MailService) renderVerifyEmailTemplate(username, domain, code string) (string, error) {
	templatePath := filepath.Join("internal", "templates", "verify_email.html")

	template, err := template.New("verify_email.html").ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("Error parsing email template: %w", err)
	}

	url := fmt.Sprintf("%s/verify/%s", strings.TrimRight(domain, "/"), code)
	data := struct {
		Username string
		Url      string
		Code     string
	}{Username: username, Url: url, Code: code}

	buffer := new(bytes.Buffer)
	if err := template.Execute(buffer, data); err != nil {
		slog.Error("Failed to execute", "err", err)
		return "", err
	}
	return buffer.String(), nil
}
