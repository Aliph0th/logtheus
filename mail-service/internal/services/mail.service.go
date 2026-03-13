package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"logtheus/mail/internal/config"
	"logtheus/mail/internal/types"
	"path/filepath"
	"strings"

	"github.com/wneessen/go-mail"
)

type MailService struct {
	client     *mail.Client
	fromHeader string
	domain     string
}

func NewMailService(cfg *config.AppConfig) *MailService {
	client, err := mail.NewClient(
		cfg.SMTP.Host,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(cfg.SMTP.Username),
		mail.WithPassword(cfg.SMTP.Password),
		mail.WithPort(cfg.SMTP.Port),
	)
	slog.Info("Initialized mail client", "host", cfg.SMTP.Host)
	if err != nil {
		panic(err)
	}
	return &MailService{client: client, fromHeader: cfg.SMTP.From, domain: strings.TrimRight(cfg.AppDomain, "/")}
}

func (s *MailService) SendVerifyEmail(to, username, code string, expiresMinutes uint8) error {
	url := fmt.Sprintf("%s/verify/%s", s.domain, code)
	data := &types.VerifyEmailData{Username: username, Url: url, Code: code, ExpiresIn: expiresMinutes}

	body, err := s.renderTemplate("verify_email.html", data)
	if err != nil {
		return err
	}
	return s.sendMail(to, "Logtheus email verification", body)
}

func (s *MailService) SendInviteEmail(to, inviteeName, referrer, projectName, code string, expiresMinutes uint8) error {
	url := fmt.Sprintf("%s/accept-invite/%s", s.domain, code)
	data := &types.InviteEmailData{InviteeName: inviteeName, Referrer: referrer, ProjectName: projectName, InviteLink: url, ExpiresIn: expiresMinutes}

	body, err := s.renderTemplate("invite_member.html", data)
	if err != nil {
		return err
	}
	return s.sendMail(to, fmt.Sprintf("You're invited to join %s on Logtheus", projectName), body)
}

func (s *MailService) renderTemplate(templateName string, data interface{}) (string, error) {
	templatePath := filepath.Join("internal", "templates", templateName)

	template, err := template.New(templateName).ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("Error parsing email template: %w", err)
	}

	buffer := new(bytes.Buffer)
	if err := template.Execute(buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
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
