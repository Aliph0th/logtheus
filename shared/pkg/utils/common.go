package utils

import (
	"logtheus/shared/pkg/consts"
	projectProto "logtheus/shared/pkg/pb/v1/project"
)

func HttpRoleToGRPCRole(httpRole consts.ProjectRole) projectProto.Role {
	switch httpRole {
	case consts.PROJECT_ROLE_OWNER:
		return projectProto.Role_OWNER
	case consts.PROJECT_ROLE_MEMBER:
		return projectProto.Role_MEMBER
	case consts.PROJECT_ROLE_VIEWER:
		return projectProto.Role_VIEWER
	default:
		return projectProto.Role_VIEWER
	}
}

func GRPCRoleToHttpRole(grpcRole projectProto.Role) consts.ProjectRole {
	switch grpcRole {
	case projectProto.Role_OWNER:
		return consts.PROJECT_ROLE_OWNER
	case projectProto.Role_MEMBER:
		return consts.PROJECT_ROLE_MEMBER
	case projectProto.Role_VIEWER:
		return consts.PROJECT_ROLE_VIEWER
	default:
		return consts.PROJECT_ROLE_VIEWER
	}
}
