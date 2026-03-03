package services

import (
	"context"
	"fmt"
	"logtheus/project/internal/config"
	"logtheus/project/internal/repository"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	mailProto "logtheus/shared/pkg/pb/v1/mail"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/shared/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InvitesService struct {
	repo           *repository.InvitesRepository
	projectService *ProjectService
	userClient     userProto.UserServiceClient
	mailClient     mailProto.MailServiceClient
	cfg            *config.AppConfig
}

func NewInvitesService(
	repo *repository.InvitesRepository,
	projectService *ProjectService,
	userClient userProto.UserServiceClient,
	mailClient mailProto.MailServiceClient,
	cfg *config.AppConfig,
) *InvitesService {
	return &InvitesService{
		repo:           repo,
		projectService: projectService,
		userClient:     userClient,
		mailClient:     mailClient,
		cfg:            cfg,
	}
}

func (s *InvitesService) CreateInvite(ctx context.Context, req *projectProto.InviteUserRequest) error {
	auth := utils.MustUserData(ctx)
	project, err := s.projectService.getByID(req.ProjectId)
	if err != nil {
		return err
	}
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	referrer, err := s.userClient.GetUser(grpcCtx, &userProto.GetUserRequest{
		Identifier: &userProto.GetUserRequest_UserId{UserId: auth.UserID},
	})
	if err != nil {
		return err
	}

	role, _ := s.projectService.GetMemberRole(referrer.Id, project.ID)
	if *role != consts.PROJECT_ROLE_OWNER && *role != consts.PROJECT_ROLE_MEMBER {
		return grpc.WithPermissionDenied("You do not have permission to invite members to this project")
	}
	count, _ := s.projectService.CountMembers(project.ID)
	maxMembers := s.cfg.Settings.MaxMembersPerProject
	if count >= maxMembers {
		err := grpc.WithResourceExhausted(fmt.Sprintf("Project '%s' has reached the maximum number of %d members", project.Name, maxMembers))
		return err.WithSlug(consts.ERROR_CODE_TOO_MANY_MEMBERS)
	}

	invitee, err := s.userClient.GetUser(grpcCtx, &userProto.GetUserRequest{
		Identifier: &userProto.GetUserRequest_Email{Email: req.Email},
	})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}
	}
	inviteeName := ""
	if invitee != nil {
		inviteeName = invitee.Username
	}
	token, err := s.userClient.IssueInviteToken(grpcCtx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	_, err = s.sendInviteEmail(inviteeName, req.Email, referrer.Username, project.Name, token.Token)
	if err != nil {
		return err
	}
	return nil
}

func (s *InvitesService) sendInviteEmail(inviteeName, inviteeEmail, referrer, projectName, token string) (*mailProto.SuccessfulResponse, error) {
	ctx := context.Background()

	req := &mailProto.SendInviteEmailRequest{
		InviteeName: inviteeName,
		Referrer:    referrer,
		ProjectName: projectName,
		Email:       inviteeEmail,
		Code:        token,
	}

	response, err := s.mailClient.SendInviteEmail(ctx, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
