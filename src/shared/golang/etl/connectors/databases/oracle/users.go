package oracle

import (
	"database/sql"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"time"
)

const permissionCte = `
	SELECT
		col.GRANTEE,
		'COLUMN::' || col.TABLE_NAME || '.' || col.COLUMN_NAME AS OBJECT,
		col.PRIVILEGE
	FROM DBA_COL_PRIVS AS col

	UNION

	SELECT
		tab.GRANTEE,
		'TABLE::' || col.TABLE_NAME AS OBJECT,
		tab.PRIVILEGE
	FROM DBA_TAB_PRIVS AS tab

	UNION

	SELECT
		sys.GRANTEE
		'SYSTEM' AS OBJECT,
		sys.PRIVILEGE
	FROM DBA_SYS_PRIVS AS sys
`

type oraclePrivilege struct {
	Grantee   sql.NullString `db:"GRANTEE"`
	Object    sql.NullString `db:"OBJECT"`
	Privilege sql.NullString `db:"PRIVILEGE"`
}

type oracleUser struct {
	Username   string    `db:"USERNAME"`
	Created    time.Time `db:"CREATED"`
	Privileges []oraclePrivilege
}

func (u oracleUser) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username:    u.Username,
		CreatedTime: &u.Created,
	}
}

type oracleRole struct {
	Role       string `db:"ROLE"`
	Privileges []oraclePrivilege
}

func (r oracleRole) toEtlRole() *types.EtlRole {
	return &types.EtlRole{
		Name: r.Role,
	}
}

type oracleRolePriv struct {
	Grantee     string `db:"GRANTEE"`
	GrantedRole string `db:"GRANTED_ROLE"`
}

type EtlOracleConnectorUser struct {
	db *databases.DB
}

func createOracleConnectorUser(db *databases.DB) (*EtlOracleConnectorUser, error) {
	return &EtlOracleConnectorUser{
		db: db,
	}, nil
}

func (c *EtlOracleConnectorUser) listDbaUsers() ([]oracleUser, *connectors.EtlSourceInfo, error) {
	src := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(fmt.Sprintf(`
		WITH perm AS (
			%s
		)
		SELECT 
			u.USERNAME,
			u.CREATED,
			perm.OBJECT,
			perm.PRIVILEGE
		FROM DBA_USERS AS u
		LEFT JOIN perm
			ON perm.GRANTEE = u.USERNAME
	`, permissionCte))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	src.AddCommand(cmd)

	allUsers := map[string]*oracleUser{}
	for rows.Next() {
		user := oracleUser{}
		priv := oraclePrivilege{}

		err = rows.Scan(&user.Username, &user.Created, &priv.Object, &priv.Privilege)
		if err != nil {
			return nil, nil, err
		}

		mapUser, ok := allUsers[user.Username]
		if !ok {
			mapUser = &user
		}
		mapUser.Privileges = append(mapUser.Privileges, priv)
		allUsers[user.Username] = mapUser
	}

	retUsers := []oracleUser{}
	for _, v := range allUsers {
		retUsers = append(retUsers, *v)
	}
	return retUsers, src, nil
}

func (c *EtlOracleConnectorUser) listDbaRoles() ([]oracleRole, *connectors.EtlSourceInfo, error) {
	src := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(fmt.Sprintf(`
		WITH perm AS (
			%s
		)
		SELECT 
			r.ROLE,
			perm.OBJECT,
			perm.PRIVILEGE
		FROM DBA_ROLES AS r
		LEFT JOIN perm
			ON perm.GRANTEE = u.USERNAME
	`, permissionCte))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	src.AddCommand(cmd)

	allRoles := map[string]*oracleRole{}
	for rows.Next() {
		role := oracleRole{}
		priv := oraclePrivilege{}

		err = rows.Scan(&role.Role, &priv.Object, &priv.Privilege)
		if err != nil {
			return nil, nil, err
		}

		mapRole, ok := allRoles[role.Role]
		if !ok {
			mapRole = &role
		}
		mapRole.Privileges = append(mapRole.Privileges, priv)
		allRoles[role.Role] = mapRole
	}

	retRoles := []oracleRole{}
	for _, v := range allRoles {
		retRoles = append(retRoles, *v)
	}
	return retRoles, src, nil
}

func (c *EtlOracleConnectorUser) listDbaRolePrivs() ([]oracleRolePriv, *connectors.EtlSourceInfo, error) {
	src := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(`
		SELECT 
			GRANTEE,
			GRANTED_ROLE
		FROM DBA_ROLE_PRIVS
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	src.AddCommand(cmd)

	allPrivs := []oracleRolePriv{}
	for rows.Next() {
		priv := oracleRolePriv{}
		err = rows.Scan(&priv.Grantee, &priv.GrantedRole)
		if err != nil {
			return nil, nil, err
		}
		allPrivs = append(allPrivs, priv)
	}

	return allPrivs, src, nil
}

func (c *EtlOracleConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	finalSource := connectors.CreateSourceInfo()

	oracleUsers, usersSrc, err := c.listDbaUsers()
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(usersSrc)

	oracleRoles, rolesSrc, err := c.listDbaRoles()
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(rolesSrc)

	roleAssignments, assignmentSrc, err := c.listDbaRolePrivs()
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(assignmentSrc)

	roleTree := CreateUserRoleTree(oracleUsers, oracleRoles, roleAssignments)

	allRoles := map[string]*types.EtlRole{}
	for _, r := range oracleRoles {
		etlRole := r.toEtlRole()
		allRoles[etlRole.Name] = etlRole
	}

	retUsers := []*types.EtlUser{}
	for _, u := range oracleUsers {
		etlUser := u.toEtlUser()

		roleNames := roleTree.FindUserParentRoleNames(u.Username)
		for _, rn := range roleNames {
			etlUser.Roles[rn] = allRoles[rn]
		}

		retUsers = append(retUsers, etlUser)
	}
	return retUsers, finalSource, nil
}
