package devops

type ProjectMembership struct {
	Username  string `json:"username" description:"Member's username，username can uniquely identify a user"`
	ProjectId string `json:"project_id" db:"project_id" description:"the DevOps Projects which project membership belongs to"`
	Role      string `json:"role" description:"DevOps Project membership's role type. e.g. owner '"`
	Status    string `json:"status" description:"Deprecated, Status of project membership. e.g. active "`
	GrantBy   string `json:"grand_by,omitempty" description:"Username of the user who assigned the role"`
}

type ProjectMemberOperator interface {
	AddProjectMember(membership *ProjectMembership) (*ProjectMembership, error)
	UpdateProjectMember(oldMembership, newMembership *ProjectMembership) (*ProjectMembership, error)
	DeleteProjectMember(membership *ProjectMembership) (*ProjectMembership, error)
}
