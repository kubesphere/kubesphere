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

var DefaultRoles = []*Role{
	{
		Name:        ProjectOwner,
		Description: "Owner have access to do all the operations of a DevOps project and own the highest permissions as well.",
	},
	{
		Name:        ProjectMaintainer,
		Description: "Maintainer have access to manage pipeline and credential configuration in a DevOps project.",
	},
	{
		Name:        ProjectDeveloper,
		Description: "Developer is able to view and trigger the pipeline.",
	},
	{
		Name:        ProjectReporter,
		Description: "Reporter is only allowed to view the status of the pipeline.",
	},
}

var AllRoleSlice = []string{ProjectDeveloper, ProjectReporter, ProjectMaintainer, ProjectOwner}

const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

type Role struct {
	Name        string `json:"name" description:"role's name e.g. owner'"`
	Description string `json:"description" description:"role 's description'"`
}
