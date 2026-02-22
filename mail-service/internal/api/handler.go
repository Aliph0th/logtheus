package api

import (
	"context"
	"log/slog"

	"logtheus/mail/internal/services"
	mailProto "logtheus/shared/pkg/pb/v1/mail"
)

type MailHandler struct {
	mailProto.UnimplementedMailServiceServer
	mailService *services.MailService
}

func NewMailHandler(mailService *services.MailService) *MailHandler {
	return &MailHandler{
		mailService: mailService,
	}
}

func (h *MailHandler) SendVerifyEmail(ctx context.Context, req *mailProto.SendVerifyEmailRequest) (*mailProto.SuccessfulResponse, error) {
	go func() {
		err := h.mailService.SendVerifyEmail(req.Email, req.Username, req.Code, uint8(req.Expiration.AsDuration().Minutes()))
		if err != nil {
			slog.Error("Failed to send verify email", "email", req.Email, "error", err)
		}
	}()
	return &mailProto.SuccessfulResponse{Queued: true}, nil
}

func (h *MailHandler) SendInviteEmail(ctx context.Context, req *mailProto.SendInviteEmailRequest) (*mailProto.SuccessfulResponse, error) {
	go func() {
		err := h.mailService.SendInviteEmail(req.Email, req.InviterName, req.ProjectName, req.Code)
		if err != nil {
			slog.Error("Failed to send invite email", "email", req.Email, "error", err)
		}
	}()
	return &mailProto.SuccessfulResponse{Queued: true}, nil
}
