package consts

type ProjectRole string

const (
	PROJECT_ROLE_OWNER  ProjectRole = "owner"
	PROJECT_ROLE_MEMBER ProjectRole = "member"
	PROJECT_ROLE_VIEWER ProjectRole = "viewer"
)

const (
	MAX_PROJECTS_PER_USER   = 5
	MAX_MEMBERS_PER_PROJECT = 20
)
