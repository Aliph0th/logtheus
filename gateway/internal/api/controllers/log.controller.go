package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/consts"
	"logtheus/gateway/internal/utils"
	ingestionProto "logtheus/shared/pkg/pb/v1/ingestion"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	ingestionClient ingestionProto.IngestionServiceClient
	logEngineClient logEngineProto.LogEngineServiceClient
}

func NewLogController(
	ingestionClient ingestionProto.IngestionServiceClient,
	logEngineClient logEngineProto.LogEngineServiceClient,
) *LogController {
	return &LogController{
		ingestionClient: ingestionClient,
		logEngineClient: logEngineClient,
	}
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

func (c *LogController) GetVolumeSeries(ctx *gin.Context) {
	data := utils.MustDTO[*dto.LogMetricsVolumeQuery](ctx)
	request, err := data.ToProtoRequest()
	if err != nil {
		excepts.RespondError(ctx, excepts.WithBadRequest("invalid query parameters"))
		return
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.GetVolumeSeries(grpcCtx, request)
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	points := make([]gin.H, 0, len(response.Points))
	for _, point := range response.Points {
		if point.Timestamp == nil {
			continue
		}
		points = append(points, gin.H{
			"timestamp": point.Timestamp.AsTime().UTC().Format(time.RFC3339),
			"count":     point.Count,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"points": points})
}

func (c *LogController) GetAggregationByField(ctx *gin.Context) {
	data := utils.MustDTO[*dto.LogMetricsAggregationQuery](ctx)
	request, err := data.ToProtoRequest()
	if err != nil {
		excepts.RespondError(ctx, excepts.WithBadRequest("invalid query parameters"))
		return
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.GetAggregationByField(grpcCtx, request)
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"items": response.Items})
}

func (c *LogController) GetLatencyStats(ctx *gin.Context) {
	data := utils.MustDTO[*dto.LogMetricsQuery](ctx)
	request, err := data.ToProtoLatencyRequest()
	if err != nil {
		excepts.RespondError(ctx, excepts.WithBadRequest("invalid query parameters"))
		return
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.GetLatencyStats(grpcCtx, request)
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"stats": response.Stats})
}
