package types

import (
	"time"
)

type EtlRole struct {
	Name        string
	Permissions map[string][]string
}

type EtlUser struct {
	Username       string
	FullName       string
	Email          string
	CreatedTime    *time.Time
	LastChangeTime *time.Time
	Roles          map[string]*EtlRole
}
