package controllers

import (
	"logtheus/internal/service"

	"github.com/gin-gonic/gin"
)

type InviteController struct {
	invitesService *service.InvitesService
}

func NewInvitesController(invitesService *service.InvitesService) *InviteController {
	return &InviteController{invitesService}
}

func (c *InviteController) CreateInvite(ctx *gin.Context) {

}
