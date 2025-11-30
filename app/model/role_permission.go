package model

type RolePermission struct {
	RoleID       string `db:"role_id"`
	PermissionID string `db:"permission_id"`
}
