package services

import (
	"context"
	"encoding/json"
	"logtheus/logengine/internal/config"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/repository"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"time"

	"github.com/pgvector/pgvector-go"
)

type LogFeatureService struct {
	cfg         *config.AppConfig
	repo        *repository.LogFeatureRepository
	logIdentity *LogIdentityService
}

func NewLogFeatureService(cfg *config.AppConfig, repo *repository.LogFeatureRepository, logIdentity *LogIdentityService) *LogFeatureService {
	return &LogFeatureService{cfg: cfg, repo: repo, logIdentity: logIdentity}
}

func (s *LogFeatureService) SaveFeatures(ctx context.Context, req *logEngineProto.SaveLogFeaturesRequest) error {
	if len(req.Features) == 0 {
		return grpc.WithInvalidArgument("Batch must contain at least one feature").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
	}

	records := make([]*models.LogFeature, 0, len(req.Features))
	for _, feature := range req.Features {
		//TODO: validate?
		attributesJSON, err := json.Marshal(feature.Attributes)
		if err != nil {
			return grpc.WithInvalidArgument("invalid attributes payload").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
		}

		createdAt := time.Now().UTC()
		if feature.ReceivedAt != nil {
			createdAt = feature.ReceivedAt.AsTime().UTC()
		}

		logID := s.logIdentity.BuildLogIDFromRawHash(
			feature.ProjectId,
			feature.ApplicationId,
			feature.SourceIp,
			feature.RawDataSha256,
		)

		records = append(records, &models.LogFeature{
			LogID:         logID,
			ApplicationID: feature.ApplicationId,
			ProjectID:     feature.ProjectId,
			Embedding:     pgvector.NewVector(feature.Embedding),
			SimilarCount:  0,
			Attributes:    json.RawMessage(attributesJSON),
			CreatedAt:     createdAt,
		})
	}

	if err := s.repo.UpsertBatch(ctx, records); err != nil {
		return grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	return nil
}
