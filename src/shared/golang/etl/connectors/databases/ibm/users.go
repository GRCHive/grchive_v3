package ibm

import (
	"database/sql"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/strings"
	"strings"
)

const (
	IBM_AUTHID_USER  string = "U"
	IBM_AUTHID_ROLE         = "R"
	IBM_AUTHID_GROUP        = "G"
)

type ibmPrivilege struct {
	Privilege sql.NullString
	Object    sql.NullString
}

type ibmPrivilegeArray []ibmPrivilege
type ibmAuthId struct {
	AuthId     string
	Privileges ibmPrivilegeArray
}

func (arr ibmPrivilegeArray) toPermissionMap() types.PermissionMap {
	ret := types.PermissionMap{}
	for _, v := range arr {
		if !v.Privilege.Valid || !v.Object.Valid {
			continue
		}

		perms, ok := ret[v.Object.String]
		if !ok {
			perms = []string{}
		}
		perms = append(perms, v.Privilege.String)
		ret[v.Object.String] = perms
	}
	return ret
}

func (id ibmAuthId) toEtlUser() *types.EtlUser {
	selfRole := id.toEtlRole()
	return &types.EtlUser{
		Username: id.AuthId,
		Roles: map[string]*types.EtlRole{
			selfRole.Name: selfRole,
		},
	}
}

func (id ibmAuthId) toEtlRole() *types.EtlRole {
	return &types.EtlRole{
		Name:        id.AuthId,
		Permissions: id.Privileges.toPermissionMap(),
	}
}

type EtlIBMConnectorUser struct {
	db *databases.DB
}

func createIBMConnectorUser(db *databases.DB) (*EtlIBMConnectorUser, error) {
	return &EtlIBMConnectorUser{
		db: db,
	}, nil
}

func (c *EtlIBMConnectorUser) getAuthIds(types ...string) ([]ibmAuthId, *connectors.EtlSourceInfo, error) {
	src := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(fmt.Sprintf(`
		WITH privs AS (
			SELECT
				AUTHID,
				PRIVILEGE,
				OBJECTTYPE CONCAT '::' CONCAT OBJECTSCHEMA CONCAT '.' OBJECTNAME AS OBJECT
			FROM SYSIBMADM.PRIVILEGES
			UNION
			SELECT
				GRANTEE AS AUTHID,
				CASE PRIVTYPE
					WHEN 'R' THEN 'REFERENCE'
					WHEN 'U' THEN 'UPDATE'
				END AS PRIVILEGE,
				'COLUMN::' CONCAT TABSCHEMA CONCAT '.' CONCAT TABNAME CONCAT '.' CONCAT COLNAME
			FROM SYSCAT.COLAUTH
		)
		SELECT
			auth.AUTHID
		FROM SYSIBMADM.AUTHORIZATIONIDS auth
		LEFT JOIN privs
			ON privs.AUTHID = auth.AUTHID
		WHERE auth.AUTHIDTYPE IN (%s)
	`, strings.Join(strings_utility.Map(types, func(s string) string {
		return fmt.Sprintf("'%s'", s)
	}), ",")))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	src.AddCommand(cmd)

	allAuths := map[string]*ibmAuthId{}
	for rows.Next() {
		auth := ibmAuthId{}
		priv := ibmPrivilege{}

		err = rows.Scan(&auth.AuthId)
		if err != nil {
			return nil, nil, err
		}

		modAuth, ok := allAuths[auth.AuthId]
		if !ok {
			modAuth = &auth
		}
		modAuth.Privileges = append(modAuth.Privileges, priv)
		allAuths[auth.AuthId] = modAuth
	}

	retAuths := []ibmAuthId{}
	for _, auth := range allAuths {
		retAuths = append(retAuths, *auth)
	}

	return retAuths, src, nil
}

func (c *EtlIBMConnectorUser) getParentGroupRoleAuthIds(authId string, authType string) ([]string, *connectors.EtlSourceInfo, error) {
	src := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(fmt.Sprintf(`
		SELECT 
			GROUP AS AUTHID
		FROM TABLE (SYSPROC.AUTH_LIST_GROUPS_FOR_AUTHID('%s')) AS T
		UNION
		SELECT 
			ROLENAME AS AUTHID
		FROM TABLE (SYSPROC.AUTH_LIST_ROLES_FOR_AUTHID('%s', '%s')) AS T
	`, authId, authId, authType))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	src.AddCommand(cmd)

	retAuths := []string{}
	for rows.Next() {
		auth := ""
		err = rows.Scan(&auth)
		if err != nil {
			return nil, nil, err
		}
		retAuths = append(retAuths, auth)
	}

	return retAuths, src, nil
}

func (c *EtlIBMConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	finalSource := connectors.CreateSourceInfo()

	// Get users. Then get groups & roles.
	// For every user, run a query to get all its parent groups and roles.
	ibmUsers, userSrc, err := c.getAuthIds(IBM_AUTHID_USER)
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(userSrc)

	allUsers := map[string]*types.EtlUser{}
	for _, u := range ibmUsers {
		user := u.toEtlUser()
		allUsers[user.Username] = user
	}

	ibmGroupsRoles, roleSrc, err := c.getAuthIds(IBM_AUTHID_ROLE, IBM_AUTHID_GROUP)
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(roleSrc)

	allRoles := map[string]*types.EtlRole{}
	for _, gr := range ibmGroupsRoles {
		role := gr.toEtlRole()
		allRoles[role.Name] = role
	}

	for _, u := range ibmUsers {
		parentGroupRoles, src, err := c.getParentGroupRoleAuthIds(u.AuthId, IBM_AUTHID_USER)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(src)

		for _, gr := range parentGroupRoles {
			role, ok := allRoles[gr]
			if !ok {
				continue
			}
			allUsers[u.AuthId].Roles[role.Name] = role
		}
	}

	retUsers := []*types.EtlUser{}
	for _, u := range allUsers {
		retUsers = append(retUsers, u)
	}

	return retUsers, finalSource, nil
}
