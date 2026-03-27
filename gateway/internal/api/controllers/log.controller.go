package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/consts"
	"logtheus/gateway/internal/utils"
	ingestionProto "logtheus/shared/pkg/pb/v1/ingestion"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	ingestionClient ingestionProto.IngestionServiceClient
}

func NewLogController(ingestionClient ingestionProto.IngestionServiceClient) *LogController {
	return &LogController{ingestionClient: ingestionClient}
}

func (c *LogController) IngestLogs(ctx *gin.Context) {
	data := utils.MustDTO[*dto.IngestLogsRequest](ctx)

	items := make([]*ingestionProto.IngestLogItem, 0, len(data.Logs))
	for _, item := range data.Logs {
		grpcItem := &ingestionProto.IngestLogItem{
			RawData: []byte(item),
		}

		items = append(items, grpcItem)
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.ingestionClient.IngestLogs(grpcCtx, &ingestionProto.IngestLogRequest{
		ApiKey:   ctx.GetString(consts.API_KEY),
		Logs:     items,
		SourceIp: ctx.ClientIP(),
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"success":        response.Success,
		"accepted_count": response.AcceptedCount,
	})
}
