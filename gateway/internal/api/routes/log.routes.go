package routes

import (
	"logtheus/gateway/internal/api/controllers"
	"logtheus/gateway/internal/api/dto"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/api/validators"
	appProto "logtheus/shared/pkg/pb/v1/application"
	"logtheus/shared/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func RegisterLogRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.LogController](container)
	appClient := utils.MustResolve[appProto.ApplicationServiceClient](container)
	authMiddleware := utils.MustResolve[gin.HandlerFunc](container)
	apiKeyMiddleware := middleware.ApiKeyMiddleware(appClient)

	logs := api.Group("/logs")
	{
		logs.Use(apiKeyMiddleware)

		logs.POST("/ingest", append(
			validators.IngestLogsValidators,
			middleware.BindDTO[*dto.IngestLogsRequest](),
			controller.IngestLogs,
		)...)
	}

	metrics := api.Group("/logs/metrics")
	{
		metrics.Use(authMiddleware)

		metrics.GET("/volume", append(
			append(validators.LogMetricsCommonValidators, validators.LogMetricsVolumeValidators...),
			middleware.ValidationMiddleware,
			middleware.BindDTO[*dto.LogMetricsVolumeQuery](),
			controller.GetVolumeSeries,
		)...)

		metrics.GET("/aggregation", append(
			append(validators.LogMetricsCommonValidators, validators.LogMetricsAggregationValidators...),
			middleware.ValidationMiddleware,
			middleware.BindDTO[*dto.LogMetricsAggregationQuery](),
			controller.GetAggregationByField,
		)...)

		metrics.GET("/latency", append(
			validators.LogMetricsCommonValidators,
			middleware.ValidationMiddleware,
			middleware.BindDTO[*dto.LogMetricsQuery](),
			controller.GetLatencyStats,
		)...)
	}
}
