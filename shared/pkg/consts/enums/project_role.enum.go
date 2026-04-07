package enums

type ProjectRole string

const (
	PROJECT_ROLE_OWNER  ProjectRole = "owner"
	PROJECT_ROLE_MEMBER ProjectRole = "member"
	PROJECT_ROLE_VIEWER ProjectRole = "viewer"
)
