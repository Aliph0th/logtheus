package utils

import (
	"context"
	"logtheus/shared/pkg/grpc"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	sharedUtils "logtheus/shared/pkg/utils"
)

func EnsureProjectReadAccess(ctx context.Context, projectID uint64, projectClient projectProto.ProjectServiceClient) error {
	_, err := getProjectRole(ctx, projectID, projectClient)
	return err
}

func EnsureProjectWriteAccess(ctx context.Context, projectID uint64, projectClient projectProto.ProjectServiceClient) error {
	role, err := getProjectRole(ctx, projectID, projectClient)
	if err != nil {
		return err
	}
	if role == projectProto.Role_VIEWER {
		return grpc.WithPermissionDenied("Insufficient permissions")
	}
	return nil
}

func getProjectRole(ctx context.Context, projectID uint64, projectClient projectProto.ProjectServiceClient) (projectProto.Role, error) {
	auth := sharedUtils.MustUserData(ctx)
	if auth == nil {
		return 0, grpc.WithUnauthenticated("Unauthorized")
	}

	roleResp, err := projectClient.GetMemberRole(ctx, &projectProto.GetMemberRoleRequest{
		UserId:    auth.UserID,
		ProjectId: projectID,
	})
	if err != nil || roleResp == nil {
		return 0, grpc.WithNotFound("Project not found or you are not a member")
	}

	return roleResp.Role, nil
}
