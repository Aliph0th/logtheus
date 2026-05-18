package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"logtheus/logengine/internal/consts"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/repository"
	"logtheus/logengine/internal/utils"
	sharedConsts "logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ClusteringService struct {
	clusteringRepo *repository.ClusteringJobRepository
	featureRepo    *repository.LogFeatureRepository
	logRepo        *repository.LogRepository
	s3             *S3Service
	logIdentity    *LogIdentityService
	projectClient  projectProto.ProjectServiceClient
}

func NewClusteringService(
	clusteringRepo *repository.ClusteringJobRepository,
	featureRepo *repository.LogFeatureRepository,
	logRepo *repository.LogRepository,
	s3 *S3Service,
	logIdentity *LogIdentityService,
	projectClient projectProto.ProjectServiceClient,
) *ClusteringService {
	return &ClusteringService{
		clusteringRepo: clusteringRepo,
		featureRepo:    featureRepo,
		logRepo:        logRepo,
		s3:             s3,
		logIdentity:    logIdentity,
		projectClient:  projectClient,
	}
}

func (s *ClusteringService) StartClusteringJob(ctx context.Context, req *logEngineProto.StartClusteringJobRequest) (*logEngineProto.StartClusteringJobResponse, error) {
	if accessErr := utils.EnsureProjectWriteAccess(ctx, req.GetFilter().GetProjectId(), s.projectClient); accessErr != nil {
		return nil, accessErr
	}

	params := utils.NormalizeClusteringParams(req)
	clusterBy := utils.NormalizeClusterBy(req.GetClusterBy())

	now := time.Now().UTC()
	requestHash, err := utils.BuildClusteringRequestHash(req, params)
	if err != nil {
		return nil, grpc.WithInternal(err).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}

	existing, repoErr := s.clusteringRepo.FindActiveByHash(ctx, requestHash)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}

	if existing != nil && existing.ExpiresAt.After(now) {
		return &logEngineProto.StartClusteringJobResponse{
			JobId:          existing.JobID.String(),
			Status:         utils.ClusteringStatusToProto(existing.Status, existing.ExpiresAt),
			ReusedExisting: true,
		}, nil
	}

	jobID := uuid.New()
	var applicationID *uint64
	if req.GetFilter().ApplicationId != nil {
		appID := req.GetFilter().GetApplicationId()
		applicationID = &appID
	}

	job := &models.ClusteringJob{
		JobID:         jobID,
		RequestHash:   requestHash,
		ProjectID:     req.GetFilter().GetProjectId(),
		ApplicationID: applicationID,
		FromTime:      req.GetFilter().GetFrom().AsTime().UTC(),
		ToTime:        req.GetFilter().GetTo().AsTime().UTC(),
		ClusterBy:     clusterBy,
		Eps:           params.Eps,
		MinPoints:     params.MinPoints,
		MaxPoints:     params.MaxPoints,
		Status:        consts.CLUSTERING_STATUS_QUEUED,
		ExpiresAt:     now.Add(time.Duration(params.TTLHours) * time.Hour),
	}

	if repoErr := s.clusteringRepo.Create(ctx, job); repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	go s.executeClusteringJob(context.Background(), jobID)

	return &logEngineProto.StartClusteringJobResponse{
		JobId:          job.JobID.String(),
		Status:         logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_QUEUED,
		ReusedExisting: false,
	}, nil
}

func (s *ClusteringService) GetClusteringJobStatus(ctx context.Context, req *logEngineProto.GetClusteringJobStatusRequest) (*logEngineProto.GetClusteringJobStatusResponse, error) {
	jobIDStr := strings.TrimSpace(req.GetJobId())
	jobID, _ := uuid.Parse(jobIDStr)

	job, repoErr := s.clusteringRepo.GetByID(ctx, jobID)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}
	if job == nil {
		return nil, grpc.WithNotFound("clustering job not found")
	}
	if accessErr := utils.EnsureProjectReadAccess(ctx, job.ProjectID, s.projectClient); accessErr != nil {
		return nil, accessErr
	}

	var startedAt, finishedAt *timestamppb.Timestamp
	if job.StartedAt != nil {
		startedAt = timestamppb.New((*job.StartedAt).UTC())
	}
	if job.FinishedAt != nil {
		finishedAt = timestamppb.New((*job.FinishedAt).UTC())
	}

	return &logEngineProto.GetClusteringJobStatusResponse{
		JobId:           job.JobID.String(),
		Status:          utils.ClusteringStatusToProto(job.Status, job.ExpiresAt),
		ProgressPercent: job.ProgressPercent,
		TotalPoints:     job.TotalPoints,
		ClusterCount:    job.ClusterCount,
		NoiseCount:      job.NoiseCount,
		ErrorMessage:    job.ErrorMessage,
		CreatedAt:       timestamppb.New(job.CreatedAt),
		StartedAt:       startedAt,
		FinishedAt:      finishedAt,
		ExpiresAt:       timestamppb.New(job.ExpiresAt),
	}, nil
}

func (s *ClusteringService) GetClusteringJobs(ctx context.Context, req *logEngineProto.GetClusteringJobsRequest) (*logEngineProto.GetClusteringJobsResponse, error) {
	projectID := req.GetProjectId()
	if projectID == 0 {
		return nil, grpc.WithInvalidArgument("project_id is required")
	}
	if accessErr := utils.EnsureProjectReadAccess(ctx, projectID, s.projectClient); accessErr != nil {
		return nil, accessErr
	}

	var applicationID *uint64
	if req.ApplicationId != nil {
		value := req.GetApplicationId()
		if value > 0 {
			applicationID = &value
		}
	}

	limit := req.GetLimit()
	if limit == 0 {
		limit = consts.DEFAULT_CLUSTERING_RESULT_PAGE_SIZE
	}
	if limit > consts.MAX_CLUSTERING_RESULT_PAGE_SIZE {
		limit = consts.MAX_CLUSTERING_RESULT_PAGE_SIZE
	}

	jobs, total, repoErr := s.clusteringRepo.ListJobs(ctx, projectID, applicationID, req.GetOffset(), limit)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}

	items := make([]*logEngineProto.ClusteringJobItem, 0, len(jobs))
	for _, job := range jobs {
		item := &logEngineProto.ClusteringJobItem{
			JobId:           job.JobID.String(),
			Status:          utils.ClusteringStatusToProto(job.Status, job.ExpiresAt),
			ProgressPercent: job.ProgressPercent,
			TotalPoints:     job.TotalPoints,
			ClusterCount:    job.ClusterCount,
			NoiseCount:      job.NoiseCount,
			ErrorMessage:    job.ErrorMessage,
			CreatedAt:       timestamppb.New(job.CreatedAt),
			ExpiresAt:       timestamppb.New(job.ExpiresAt),
			ProjectId:       job.ProjectID,
			ClusterBy:       job.ClusterBy,
			From:            timestamppb.New(job.FromTime),
			To:              timestamppb.New(job.ToTime),
		}
		if job.ApplicationID != nil {
			appID := *job.ApplicationID
			item.ApplicationId = &appID
		}
		if job.StartedAt != nil {
			item.StartedAt = timestamppb.New(job.StartedAt.UTC())
		}
		if job.FinishedAt != nil {
			item.FinishedAt = timestamppb.New(job.FinishedAt.UTC())
		}
		items = append(items, item)
	}

	return &logEngineProto.GetClusteringJobsResponse{
		Jobs:  items,
		Total: total,
	}, nil
}

func (s *ClusteringService) GetClusteringJobResult(ctx context.Context, req *logEngineProto.GetClusteringJobResultRequest) (*logEngineProto.GetClusteringJobResultResponse, error) {
	jobIDStr := strings.TrimSpace(req.GetJobId())
	jobID, _ := uuid.Parse(jobIDStr)

	job, repoErr := s.clusteringRepo.GetByID(ctx, jobID)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}
	if job == nil {
		return nil, grpc.WithNotFound("clustering job not found")
	}
	if accessErr := utils.EnsureProjectReadAccess(ctx, job.ProjectID, s.projectClient); accessErr != nil {
		return nil, accessErr
	}
	if time.Now().UTC().After(job.ExpiresAt) {
		return nil, grpc.WithNotFound("clustering job result expired").WithSlug(sharedConsts.ERROR_CODE_CLUSTERING_EXPIRED)
	}
	if job.Status != consts.CLUSTERING_STATUS_SUCCEEDED {
		return nil, grpc.WithInvalidArgument("clustering job is not finished yet").WithSlug(sharedConsts.ERROR_CODE_CLUSTERING_NOT_READY)
	}

	limit := req.GetLimit()
	if limit == 0 {
		limit = consts.DEFAULT_CLUSTERING_RESULT_PAGE_SIZE
	}
	if limit > consts.MAX_CLUSTERING_RESULT_PAGE_SIZE {
		limit = consts.MAX_CLUSTERING_RESULT_PAGE_SIZE
	}
	offset := req.GetOffset()

	rows, totalItems, repoErr := s.clusteringRepo.GetAssignmentsPaged(ctx, jobID, offset, limit)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}

	summaries, repoErr := s.clusteringRepo.GetClusterSummaries(ctx, jobID)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}

	assignments := make([]*logEngineProto.ClusteringAssignmentItem, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, &logEngineProto.ClusteringAssignmentItem{
			LogId:     row.LogID,
			ClusterId: row.ClusterID,
		})
	}

	sampleTextByClusterID := s.buildClusterSampleTexts(ctx, jobID)

	clusters := make([]*logEngineProto.ClusteringClusterSummaryItem, 0, len(summaries))
	for _, summary := range summaries {
		clusters = append(clusters, &logEngineProto.ClusteringClusterSummaryItem{
			ClusterId:  summary.ClusterID,
			Size:       summary.Size,
			SampleText: sampleTextByClusterID[summary.ClusterID],
		})
	}

	return &logEngineProto.GetClusteringJobResultResponse{
		JobId:       jobID.String(),
		Status:      logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_SUCCEEDED,
		TotalItems:  totalItems,
		Assignments: assignments,
		Clusters:    clusters,
	}, nil
}

func (s *ClusteringService) CancelClusteringJob(ctx context.Context, req *logEngineProto.CancelClusteringJobRequest) (*logEngineProto.CancelClusteringJobResponse, error) {
	jobIDStr := strings.TrimSpace(req.GetJobId())
	jobID, _ := uuid.Parse(jobIDStr)

	job, repoErr := s.clusteringRepo.GetByID(ctx, jobID)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}
	if job == nil {
		return nil, grpc.WithNotFound("clustering job not found")
	}
	if accessErr := utils.EnsureProjectWriteAccess(ctx, job.ProjectID, s.projectClient); accessErr != nil {
		return nil, accessErr
	}
	job, repoErr = s.clusteringRepo.GetByID(ctx, jobID)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}
	if job == nil {
		return nil, grpc.WithNotFound("clustering job not found")
	}

	canceled, repoErr := s.clusteringRepo.MarkCanceled(ctx, jobID, time.Now().UTC())
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(sharedConsts.INTERNAL_ERROR_CODE_CLUSTERING_FAILED)
	}

	return &logEngineProto.CancelClusteringJobResponse{
		Canceled: canceled,
		Status:   utils.ClusteringStatusToProto(job.Status, job.ExpiresAt),
	}, nil
}

func (s *ClusteringService) executeClusteringJob(ctx context.Context, jobID uuid.UUID) {
	now := time.Now().UTC()
	running, err := s.clusteringRepo.MarkRunning(ctx, jobID, now)
	if err != nil {
		slog.Error("Failed to mark clustering job as running", "job_id", jobID, "error", err)
		return
	}
	if !running {
		return
	}

	job, err := s.clusteringRepo.GetByID(ctx, jobID)
	if err != nil || job == nil {
		slog.Error("Failed to load clustering job", "job_id", jobID, "error", err)
		return
	}

	if err := s.clusteringRepo.UpdateProgress(ctx, jobID, 5, 0); err != nil {
		slog.Warn("Failed to update clustering progress", "job_id", jobID, "error", err)
	}

	filter := &logEngineProto.ClusteringFilter{
		ProjectId: job.ProjectID,
		From:      timestamppb.New(job.FromTime),
		To:        timestamppb.New(job.ToTime),
	}
	if job.ApplicationID != nil {
		appID := *job.ApplicationID
		filter.ApplicationId = &appID
	}

	features, truncated, err := s.featureRepo.ListByFilterWithLimit(ctx, filter, job.MaxPoints)
	if err != nil {
		_ = s.clusteringRepo.MarkFailed(ctx, jobID, "failed to load embeddings", time.Now().UTC())
		slog.Error("Failed to load embeddings for clustering", "job_id", jobID, "error", err)
		return
	}
	if truncated {
		slog.Warn("Clustering dataset exceeded max_points, using truncated sample", "job_id", jobID, "max_points", job.MaxPoints)
	}

	totalPoints := uint32(len(features))
	if err := s.clusteringRepo.UpdateProgress(ctx, jobID, 20, totalPoints); err != nil {
		slog.Warn("Failed to update clustering progress", "job_id", jobID, "error", err)
	}

	if len(features) == 0 {
		if err := s.clusteringRepo.SaveSucceededResult(ctx, jobID, nil, nil, 0, 0, time.Now().UTC()); err != nil {
			slog.Error("Failed to save empty clustering result", "job_id", jobID, "error", err)
		}
		return
	}

	labels := make([]int32, 0, len(features))
	if job.ClusterBy == consts.CLUSTERING_TARGET_EMBEDDING {
		vectors := make([][]float32, len(features))
		for idx := range features {
			vectors[idx] = features[idx].Embedding.Slice()
		}
		labels = utils.RunDBSCANCosine(vectors, job.Eps, int(job.MinPoints))
	} else {
		values := make([]string, len(features))
		for idx := range features {
			if len(features[idx].Attributes) == 0 {
				continue
			}

			attributes := map[string]any{}
			if err := json.Unmarshal(features[idx].Attributes, &attributes); err != nil {
				continue
			}

			rawValue, exists := attributes[job.ClusterBy]
			if !exists || rawValue == nil {
				continue
			}

			values[idx] = strings.TrimSpace(fmt.Sprint(rawValue))
		}

		labels = utils.ClusterByExactValue(values)
	}

	if err := s.clusteringRepo.UpdateProgress(ctx, jobID, 80, totalPoints); err != nil {
		slog.Warn("Failed to update clustering progress", "job_id", jobID, "error", err)
	}

	now = time.Now().UTC()
	assignments := make([]*models.ClusteringAssignment, 0, len(labels))
	clusterSizeByID := make(map[int32]uint32)
	var noiseCount uint32

	for idx, label := range labels {
		assignments = append(assignments, &models.ClusteringAssignment{
			JobID:     jobID,
			LogID:     features[idx].LogID,
			ClusterID: label,
			CreatedAt: now,
		})

		if label == -1 {
			noiseCount++
			continue
		}
		clusterSizeByID[label]++
	}

	clusterIDs := make([]int, 0, len(clusterSizeByID))
	for clusterID := range clusterSizeByID {
		clusterIDs = append(clusterIDs, int(clusterID))
	}
	sort.Ints(clusterIDs)

	summaries := make([]*models.ClusteringClusterSummary, 0, len(clusterIDs))
	for _, clusterID := range clusterIDs {
		summaries = append(summaries, &models.ClusteringClusterSummary{
			JobID:     jobID,
			ClusterID: int32(clusterID),
			Size:      clusterSizeByID[int32(clusterID)],
			CreatedAt: now,
		})
	}

	if err := s.clusteringRepo.SaveSucceededResult(ctx, jobID, assignments, summaries, uint32(len(summaries)), noiseCount, now); err != nil {
		_ = s.clusteringRepo.MarkFailed(ctx, jobID, "failed to store clustering result", time.Now().UTC())
		slog.Error("Failed to save clustering result", "job_id", jobID, "error", err)
		return
	}
}

func (s *ClusteringService) buildClusterSampleTexts(ctx context.Context, jobID uuid.UUID) map[int32]string {
	result := map[int32]string{}
	representativeByClusterID, err := s.clusteringRepo.GetClusterRepresentativeLogIDs(ctx, jobID)
	if err != nil {
		slog.Warn("Failed to load representative log ids for cluster samples", "error", err)
		return result
	}
	if len(representativeByClusterID) == 0 {
		return result
	}

	logIDs := make([]string, 0, len(representativeByClusterID))
	for _, logID := range representativeByClusterID {
		logIDs = append(logIDs, logID)
	}

	records, err := s.logRepo.GetByIDs(ctx, logIDs)
	if err != nil {
		slog.Warn("Failed to load log records for cluster samples", "error", err)
		return result
	}

	recordByLogID := make(map[string]repository.LogRecordLight, len(records))
	for _, record := range records {
		recordByLogID[record.LogID] = record
	}

	s3PayloadCache := map[string][]byte{}
	for clusterID, logID := range representativeByClusterID {
		record, exists := recordByLogID[logID]
		if !exists || strings.TrimSpace(record.S3Key) == "" || record.S3Key == consts.SimilarAggregateS3Key {
			continue
		}

		payload, exists := s3PayloadCache[record.S3Key]
		if !exists {
			downloaded, downloadErr := s.s3.DownloadObject(ctx, record.S3Key)
			if downloadErr != nil {
				slog.Warn("Failed to download sample S3 object", "s3_key", record.S3Key, "error", downloadErr)
				continue
			}
			payload = downloaded
			s3PayloadCache[record.S3Key] = payload
		}

		sampleText := s.findSampleTextByLogID(record, payload)
		if sampleText != "" {
			result[clusterID] = sampleText
		}
	}

	return result
}

func (s *ClusteringService) findSampleTextByLogID(record repository.LogRecordLight, payload []byte) string {
	lines := bytes.Split(payload, []byte("\n"))
	for _, line := range lines {
		candidate := strings.TrimSpace(string(line))
		if candidate == "" {
			continue
		}

		candidateLogID := s.logIdentity.BuildLogIDFromRawData(
			record.ProjectID,
			record.ApplicationID,
			record.SourceIP,
			[]byte(candidate),
		)
		if candidateLogID == record.LogID {
			return candidate
		}
	}

	return ""
}
