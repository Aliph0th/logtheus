package service

import (
	"log/slog"
	"logtheus/internal/api/dto"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/config"
	"logtheus/internal/consts"
	"logtheus/internal/consts/enums"
	"logtheus/internal/repository"
	"logtheus/internal/utils"
	sl "logtheus/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

type InvitesService struct {
	repo           *repository.InviteRepository
	userService    *UserService
	mailService    *MailService
	tokenService   *TokenService
	projectService *ProjectService
	cfg            *config.AppConfig
}

func NewInvitesService(
	repo *repository.InviteRepository,
	userService *UserService,
	mailService *MailService,
	tokenService *TokenService,
	projectService *ProjectService,
	cfg *config.AppConfig,
) *InvitesService {
	return &InvitesService{
		repo:           repo,
		userService:    userService,
		mailService:    mailService,
		projectService: projectService,
		tokenService:   tokenService,
		cfg:            cfg,
	}
}

func (s *InvitesService) CreateInvite(ctx *gin.Context, dto *dto.InviteCreateRequest) error {
	auth := utils.MustAuth(ctx)
	project, err := s.projectService.GetByID(dto.ProjectID)
	if err != nil {
		return err
	}
	referrer, err := s.userService.GetUserByID(auth.UserID)
	if err != nil {
		return err
	}

	role, _ := s.projectService.GetMemberRole(referrer.ID, project.ID)
	if *role != enums.PROJECT_ROLE_OWNER && *role != enums.PROJECT_ROLE_MEMBER {
		return excepts.WithForbidden("You do not have permission to invite members to this project")
	}
	count, _ := s.projectService.CountMembers(project.ID)
	if count >= consts.MAX_MEMBERS_PER_PROJECT {
		return excepts.WithConflict("Project member limit reached")
	}

	invitee, err := s.userService.GetUserByEmail(dto.Email)
	if err != nil {
		return err
	}
	token, err := s.tokenService.generateToken(enums.TOKEN_TYPE_INVITE)
	if err != nil {
		return err
	}
	go func() {
		err := s.mailService.SendInviteEmail(invitee.Email, referrer.Username, project.Name, s.cfg.AppDomain, token)
		if err != nil {
			slog.Error("Invite email failed", "email", invitee.Email, sl.Error(err))
		}
	}()
	return nil
}
