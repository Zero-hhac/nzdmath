package consts

const (
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
	RoleMember     = "member"

	StatusActive   = 1
	StatusDisabled = 0

	CommentStatusVisible  = 1
	CommentStatusHidden   = 0

	TargetTypeEvent    = "event"
	TargetTypeNews     = "news"
	TargetTypeResource = "resource"
	TargetTypeShowcase = "showcase"

	DefaultPageSize = 10
	MaxPageSize     = 100
	MaxUploadSize   = 50 * 1024 * 1024

	CacheTTLHomepage    = 300
	CacheTTLList        = 300
	CacheTTLDetail      = 600
	CacheTTLDashboard   = 30
)

var Departments = []string{"组织部", "办公室", "宣传部", "外联部"}

func IsValidDepartment(d string) bool {
	for _, dep := range Departments {
		if dep == d {
			return true
		}
	}
	return false
}