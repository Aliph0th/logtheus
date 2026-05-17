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

	items := response.Items
	if items == nil {
		items = []*logEngineProto.AggregationItem{}
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
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
	var statsPayload gin.H
	if stats := response.GetStats(); stats != nil {
		statsPayload = gin.H{
			"p50": utils.NormalizeFloat(stats.GetP50()),
			"p95": utils.NormalizeFloat(stats.GetP95()),
			"p99": utils.NormalizeFloat(stats.GetP99()),
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"stats": statsPayload})
}

func (c *LogController) StartClusteringJob(ctx *gin.Context) {
	data := utils.MustDTO[*dto.LogClusteringStartRequest](ctx)
	request, err := data.ToProtoRequest()
	if err != nil {
		excepts.RespondError(ctx, excepts.WithBadRequest(err.Error()))
		return
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.StartClusteringJob(grpcCtx, request)
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"job_id":          response.JobId,
		"status":          response.Status.String(),
		"reused_existing": response.ReusedExisting,
	})
}

func (c *LogController) GetClusteringJobStatus(ctx *gin.Context) {
	jobID := ctx.Param("job_id")

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.GetClusteringJobStatus(grpcCtx, &logEngineProto.GetClusteringJobStatusRequest{JobId: jobID})
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	payload := gin.H{
		"job_id":           response.JobId,
		"status":           response.Status.String(),
		"progress_percent": response.ProgressPercent,
		"total_points":     response.TotalPoints,
		"cluster_count":    response.ClusterCount,
		"noise_count":      response.NoiseCount,
		"error_message":    response.ErrorMessage,
		"created_at":       "",
		"started_at":       nil,
		"finished_at":      nil,
		"expires_at":       "",
	}

	if response.CreatedAt != nil {
		payload["created_at"] = response.CreatedAt.AsTime().UTC().Format(time.RFC3339)
	}
	if response.StartedAt != nil {
		payload["started_at"] = response.StartedAt.AsTime().UTC().Format(time.RFC3339)
	}
	if response.FinishedAt != nil {
		payload["finished_at"] = response.FinishedAt.AsTime().UTC().Format(time.RFC3339)
	}
	if response.ExpiresAt != nil {
		payload["expires_at"] = response.ExpiresAt.AsTime().UTC().Format(time.RFC3339)
	}

	ctx.JSON(http.StatusOK, payload)
}

func (c *LogController) GetClusteringJobResult(ctx *gin.Context) {
	jobID := ctx.Param("job_id")
	data := utils.MustDTO[*dto.LogClusteringResultQuery](ctx)

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.GetClusteringJobResult(grpcCtx, data.ToProtoRequest(jobID))
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"job_id":      response.JobId,
		"status":      response.Status.String(),
		"total_items": response.TotalItems,
		"assignments": response.Assignments,
		"clusters":    response.Clusters,
	})
}

func (c *LogController) GetClusteringJobs(ctx *gin.Context) {
	data := utils.MustDTO[*dto.LogClusteringJobsQuery](ctx)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.GetClusteringJobs(grpcCtx, data.ToProtoRequest())
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	jobs := make([]gin.H, 0, len(response.Jobs))
	for _, job := range response.Jobs {
		payload := gin.H{
			"job_id":           job.JobId,
			"status":           job.Status.String(),
			"progress_percent": job.ProgressPercent,
			"total_points":     job.TotalPoints,
			"cluster_count":    job.ClusterCount,
			"noise_count":      job.NoiseCount,
			"error_message":    job.ErrorMessage,
			"created_at":       "",
			"started_at":       nil,
			"finished_at":      nil,
			"expires_at":       "",
			"project_id":       job.ProjectId,
			"application_id":   nil,
			"cluster_by":       job.ClusterBy,
			"from":             "",
			"to":               "",
		}
		if job.CreatedAt != nil {
			payload["created_at"] = job.CreatedAt.AsTime().UTC().Format(time.RFC3339)
		}
		if job.StartedAt != nil {
			payload["started_at"] = job.StartedAt.AsTime().UTC().Format(time.RFC3339)
		}
		if job.FinishedAt != nil {
			payload["finished_at"] = job.FinishedAt.AsTime().UTC().Format(time.RFC3339)
		}
		if job.ExpiresAt != nil {
			payload["expires_at"] = job.ExpiresAt.AsTime().UTC().Format(time.RFC3339)
		}
		if job.ApplicationId != nil {
			payload["application_id"] = job.GetApplicationId()
		}
		if job.From != nil {
			payload["from"] = job.From.AsTime().UTC().Format(time.RFC3339)
		}
		if job.To != nil {
			payload["to"] = job.To.AsTime().UTC().Format(time.RFC3339)
		}
		jobs = append(jobs, payload)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"total": response.Total,
	})
}

func (c *LogController) CancelClusteringJob(ctx *gin.Context) {
	jobID := ctx.Param("job_id")

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, grpcErr := c.logEngineClient.CancelClusteringJob(grpcCtx, &logEngineProto.CancelClusteringJobRequest{JobId: jobID})
	if grpcErr != nil {
		excepts.RespondError(ctx, grpcErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"canceled": response.Canceled,
		"status":   response.Status.String(),
	})
}
