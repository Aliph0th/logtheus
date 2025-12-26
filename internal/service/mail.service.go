package service

import (
	"bytes"
	"fmt"
	"html/template"
	"logtheus/internal/config"
	"path/filepath"
	"strings"

	"gopkg.in/gomail.v2"
)

type MailService struct {
	dialer     *gomail.Dialer
	fromHeader string
}

func NewMailService(cfg *config.AppConfig) *MailService {
	dialer := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Login, cfg.SMTP.Password)
	return &MailService{dialer: dialer, fromHeader: cfg.SMTP.From}
}

func (s *MailService) SendVerifyEmail(to, username, domain, code string) error {
	body, err := s.renderVerifyEmailTemplate(username, domain, code)
	if err != nil {
		return err
	}
	return s.sendMail(to, "Logtheus email verification", body)
}

func (s *MailService) sendMail(to, subject, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", s.fromHeader)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	return s.dialer.DialAndSend(message)
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
		return "", err
	}
	return buffer.String(), nil
}
