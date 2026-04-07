package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"logtheus/logengine/internal/config"
	"regexp"
	"strings"
	"time"

	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

var re1 = regexp.MustCompile(`[\W_]`)
var re2 = regexp.MustCompile("-{2,}")

type S3Service struct {
	client *s3.Client
}

func NewS3Service(cfg *config.AppConfig) *S3Service {
	creds := credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, "")
	client := s3.New(s3.Options{
		Credentials:      creds,
		RetryMode:        aws.RetryModeStandard,
		BaseEndpoint:     &cfg.S3.Host,
		Region:           cfg.S3.Region,
		RetryMaxAttempts: 3,
	})
	return &S3Service{client: client}
}

func (s *S3Service) UploadBatch(ctx context.Context, logs []*logEngineProto.LogItem) (string, error) {
	if len(logs) == 0 {
		return "", fmt.Errorf("batch is empty")
	}

	first := logs[0]
	bucket := s.getCanonicalProject(first.ProjectId)
	if err := s.ensureBucketExists(ctx, bucket); err != nil {
		return "", err
	}

	folder := s.getCanonicalApplication(first.ApplicationId, first.ApplicationName)
	key := fmt.Sprintf("%s/%s.log", folder, first.ReceivedAt.AsTime().UTC().Format(time.RFC3339Nano))
	body := bytes.NewReader(mergeBatchRawData(logs))

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", bucket, key), nil
}

func (s *S3Service) ensureBucketExists(ctx context.Context, bucket string) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return err
	}

	switch apiErr.ErrorCode() {
	case "NoSuchBucket", "NotFound", "404":
		_, createErr := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
		if createErr == nil {
			return nil
		}

		var createAPIErr smithy.APIError
		if errors.As(createErr, &createAPIErr) {
			switch createAPIErr.ErrorCode() {
			case "BucketAlreadyOwnedByYou", "BucketAlreadyExists":
				return nil
			}
		}

		return createErr
	default:
		return err
	}
}

func (s *S3Service) DeleteObject(ctx context.Context, s3Key string) error {
	parts := strings.SplitN(s3Key, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid s3 key format")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(parts[0]),
		Key:    aws.String(parts[1]),
	})
	return err
}

func mergeBatchRawData(logs []*logEngineProto.LogItem) []byte {
	if len(logs) == 1 {
		return logs[0].RawData
	}

	totalSize := 0
	for _, item := range logs {
		totalSize += len(item.RawData) + 1
	}

	payload := make([]byte, 0, totalSize)
	for idx, item := range logs {
		payload = append(payload, item.RawData...)
		if idx < len(logs)-1 {
			payload = append(payload, '\n')
		}
	}

	return payload
}

func (s *S3Service) getCanonicalProject(projectID uint64) string {
	return fmt.Sprintf("project-%d", projectID)
}

func (s *S3Service) getCanonicalApplication(applicationID uint64, name string) string {
	name = re1.ReplaceAllString(name, "-")
	name = strings.Trim(re2.ReplaceAllString(name, "-"), "-")
	return fmt.Sprintf("app_%d_%s", applicationID, name)
}
