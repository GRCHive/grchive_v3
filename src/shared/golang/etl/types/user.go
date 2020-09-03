package types

import (
	"time"
)

type PermissionMap = map[string][]string

type EtlRole struct {
	Name        string
	Permissions PermissionMap
	Denied      PermissionMap
}

type EtlUser struct {
	Username       string
	FullName       string
	Email          string
	CreatedTime    *time.Time
	LastChangeTime *time.Time
	Roles          map[string]*EtlRole
	NestedUsers    map[string]*EtlUser
}
