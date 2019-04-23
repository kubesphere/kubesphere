package devops

const (
	DevOpsProjectMembershipTableName       = "project_membership"
	DevOpsProjectMembershipUsernameColumn  = "project_membership.username"
	DevOpsProjectMembershipProjectIdColumn = "project_membership.project_id"
	DevOpsProjectMembershipRoleColumn      = "project_membership.role"
)

type DevOpsProjectMembership struct {
	Username  string `json:"username"`
	ProjectId string `json:"project_id" db:"project_id"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	GrantBy   string `json:"grand_by,omitempty"`
}

var DevOpsProjectMembershipColumns = GetColumnsFromStruct(&DevOpsProjectMembership{})

func NewDevOpsProjectMemberShip(username, projectId, role, grantBy string) *DevOpsProjectMembership {
	return &DevOpsProjectMembership{
		Username:  username,
		ProjectId: projectId,
		Role:      role,
		Status:    StatusActive,
		GrantBy:   grantBy,
	}
}
