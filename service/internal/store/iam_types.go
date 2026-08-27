package store

const (
	UserCreatedByAdmin = "admin-created"
	UserCreatedBySSO   = "sso-autocreated"
)

type UserAccountRow struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedBy    string
	CreatedAt    string
	UpdatedAt    string
}

type SessionRow struct {
	ID                 int
	SID                string
	UserAccountID      int
	ImpersonatorUserID *int
	CreatedAt          string
	UpdatedAt          string
}

type APIKeyRow struct {
	ID            int
	UserAccountID int
	Name          string
	KeyValue      string
	ReadOnly      bool
	LastUsedAt    string
	CreatedAt     string
	UpdatedAt     string
}

type UserGroupRow struct {
	ID          int
	Name        string
	MemberCount int
	CreatedAt   string
	UpdatedAt   string
}

type RBACPermissionRow struct {
	ID          int
	Name        string
	Description string
}

type RBACRoleRow struct {
	ID            int
	Name          string
	Description   string
	PermissionIDs []int
	GroupCount    int
	UserCount     int
}

type MyPermissionAuditRow struct {
	Permission     string
	Granted        bool
	GrantingGroups []string
}
