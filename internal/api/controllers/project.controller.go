package controllers

import (
	"logtheus/internal/api/dto"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/consts"
	"logtheus/internal/service"
	utils "logtheus/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	projectService *service.ProjectService
}

func NewProjectController(projectService *service.ProjectService) *ProjectController {
	return &ProjectController{projectService}
}

func (c *ProjectController) CreateProject(ctx *gin.Context) {
	data := utils.MustDTO[dto.ProjectCreateRequest](ctx)

	project, err := c.projectService.CreateProject(ctx, &data)
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"project": project, "max": consts.MAX_PROJECTS_PER_USER})
}

func (c *ProjectController) GetMyProjects(ctx *gin.Context) {
	projects, err := c.projectService.GetMyProjects(ctx)
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"projects": projects, "max": consts.MAX_PROJECTS_PER_USER})
}

func (c *ProjectController) GetProjectByID(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	project, err := c.projectService.GetProjectByID(ctx, id)
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"project": project})
}
