package mssql

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"strings"
	"time"
)

const (
	SERVER_PRINCIPAL_TABLE     string = "sys.server_principals"
	SERVER_ROLE_MEMBER_TABLE          = "sys.server_role_members"
	SERVER_PERMISSIONS_TABLE          = "sys.server_permissions"
	DATABASE_PRINCIPAL_TABLE          = "sys.database_principals"
	DATABASE_ROLE_MEMBER_TABLE        = "sys.database_role_members"
	DATABASE_PERMISSIONS_TABLE        = "sys.database_permissions"
)

const (
	SERVER_PRINCIPAL_SQL_LOGIN                string = "S"
	SERVER_PRINCIPAL_WINDOWS_LOGIN                   = "U"
	SERVER_PRINCIPAL_WINDOWS_GROUP                   = "G"
	SERVER_PRINCIPAL_SERVER_ROLE                     = "R"
	DATABASE_PRINCIPAL_APPLICATION_ROLE              = "A"
	DATABASE_PRINCIPAL_EXTERNAL_USER_FROM_AD         = "E"
	DATABASE_PRINCIPAL_EXTERNAL_GROUP_FROM_AD        = "X"
	DATABASE_PRINCIPAL_DATABASE_ROLE                 = "R"
	DATABASE_PRINCIPAL_SQL_USER                      = "S"
	DATABASE_PRINCIPAL_WINDOWS_USER                  = "U"
	DATABASE_PRINCIPAL_WINDOWS_GROUP                 = "G"
)

const (
	PERMISSION_DENY_STATE              string = "D"
	PERMISSION_REVOKE_STATE                   = "R"
	PERMISSION_GRANT_STATE                    = "G"
	PERMISSION_GRANT_WITH_OPTION_STATE        = "W"
)

type mssqlPermission struct {
	PermissionName string
	State          string
	Object         string
}

type mssqlPrincipal struct {
	Name        string
	PrincipalId int32
	Sid         []byte
	CreateDate  time.Time
	Permissions []mssqlPermission
}

func (p mssqlPrincipal) toEtlUser() *types.EtlUser {
	selfRole := p.toEtlRole()
	return &types.EtlUser{
		Username:    p.Name,
		CreatedTime: &p.CreateDate,
		Roles: map[string]*types.EtlRole{
			selfRole.Name: selfRole,
		},
	}
}

func (p mssqlPrincipal) toEtlRole() *types.EtlRole {
	granted := types.PermissionMap{}
	denied := types.PermissionMap{}

	for _, perm := range p.Permissions {
		if perm.State == PERMISSION_DENY_STATE || perm.State == PERMISSION_REVOKE_STATE {
			denied[perm.Object] = append(denied[perm.Object], perm.PermissionName)
		} else {
			granted[perm.Object] = append(granted[perm.Object], perm.PermissionName)
		}
	}

	return &types.EtlRole{
		Name:        p.Name,
		Permissions: granted,
		Denied:      denied,
	}
}

type mssqlRoleMember struct {
	RolePrincipalId   int32
	MemberPrincipalId int32
}

type EtlMssqlConnectorUser struct {
	db *databases.DB
}

func createMssqlConnectorUser(db *databases.DB) (*EtlMssqlConnectorUser, error) {
	return &EtlMssqlConnectorUser{
		db: db,
	}, nil
}

// Retrieves the server principals along with their corresponding granted permissions.
func (c *EtlMssqlConnectorUser) getPrincipals(principalTable string, permissionTable string, types ...string) ([]mssqlPrincipal, *connectors.EtlSourceInfo, error) {
	source := connectors.CreateSourceInfo()
	allPrincipals := map[int32]*mssqlPrincipal{}

	rows, cmd, err := c.db.LoggedQuery(fmt.Sprintf(`
		SELECT 
			prin.name,
			prin.principal_id,
			prin.sid, 
			prin.create_date,
			perm.permission_name,
			perm.state,
			CONCAT(perm.class_desc, '::', o.type_desc, '::', s.name, '.', o.name) AS object
		FROM %s AS prin
		LEFT JOIN %s AS perm
			ON perm.grantee_principal_id = prin.principal_id
		LEFT JOIN sys.objects AS o
			ON perm.major_id = o.object_id
		LEFT JOIN sys.schemas AS s
			ON s.schema_id = o.schema_id
		WHERE type IN (%s) AND sid IS NOT NULL
	`, principalTable, permissionTable, strings.Join(types, ",")))

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	source.AddCommand(cmd)

	for rows.Next() {
		prin := mssqlPrincipal{}
		perm := mssqlPermission{}

		err = rows.Scan(
			&prin.Name,
			&prin.PrincipalId,
			&prin.Sid,
			&prin.CreateDate,
			&perm.PermissionName,
			&perm.State,
			&perm.Object,
		)
		if err != nil {
			return nil, nil, err
		}

		currentPrincipal, ok := allPrincipals[prin.PrincipalId]
		if !ok {
			currentPrincipal = &prin
		}

		currentPrincipal.Permissions = append(currentPrincipal.Permissions, perm)
		allPrincipals[prin.PrincipalId] = currentPrincipal
	}

	principals := []mssqlPrincipal{}
	for _, p := range allPrincipals {
		principals = append(principals, *p)
	}
	return principals, source, nil
}

func (c *EtlMssqlConnectorUser) getRoleMembers(table string) ([]mssqlRoleMember, *connectors.EtlSourceInfo, error) {
	source := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(fmt.Sprintf(`
		SELECT role_principal_id, member_principal_id FROM %s
	`, table))

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	source.AddCommand(cmd)

	members := []mssqlRoleMember{}
	for rows.Next() {
		m := mssqlRoleMember{}
		err = rows.Scan(&m.RolePrincipalId, &m.MemberPrincipalId)
		if err != nil {
			return nil, nil, err
		}
		members = append(members, m)
	}
	return members, source, nil
}

// Returns a mapping from the SID of the user to the EtlUser object.
func getEtlUsersAndRolesFromMssqlPrincipalsAndRoles(logins []mssqlPrincipal, roles []mssqlPrincipal, members []mssqlRoleMember) (map[string]*types.EtlUser, error) {
	sidToUser := map[string]*types.EtlUser{}

	// Need to keep track of the principal id to sid mapping to keep track of role membership.
	loginPidToSid := map[int32]string{}

	for _, u := range logins {
		sid := SidToString(u.Sid)
		sidToUser[sid] = u.toEtlUser()
		loginPidToSid[u.PrincipalId] = sid
	}

	{
		pidToRole := map[int32]*types.EtlRole{}
		for _, r := range roles {
			pidToRole[r.PrincipalId] = r.toEtlRole()
		}

		// Construct a principal tree where the nodes is a principal and each edge is defined by the connection specified in mssqlRoleMember.
		tree := ConstructPrincipalTree(append(logins, roles...), members)

		// For each login, find all the parent principal roles and add those roles to the relevant etl user.
		for _, u := range logins {
			relevantRoles := tree.FindParentPrincipalsFromId(u.PrincipalId)
			etlUser := sidToUser[SidToString(u.Sid)]
			for _, r := range relevantRoles {
				etlRole := pidToRole[r.PrincipalId]
				etlUser.Roles[etlRole.Name] = etlRole
			}
		}
	}

	return sidToUser, nil
}

func (c *EtlMssqlConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	// Step 1: Get all Server Logins (all the users who can login to the database) -> map each to an EtlUser (with a Self role)
	// Step 2: Get all Server Roles -> map each to an EtlRole
	// Step 3: Get all Server Permissions -> add permissions to each role as necessary.
	// Step 4: Get mapping from server roles to server logins
	// Step 5: Get all Database Users (along with their mapping to Server Logins) -> store as a nested EtlUser fro the server login
	// Step 6: Get all Database Roles
	// Step 7: Get all Database Permissions
	// Step 8: Get mapping from database roles to database users

	finalSource := connectors.CreateSourceInfo()

	// Map MS SQL principal SID to user. We need to use SID as the key because that's what MS SQL uses to track
	// the connection between database userse and server logins.
	sidToUser := map[string]*types.EtlUser{}

	{
		serverLogins, src, err := c.getPrincipals(SERVER_PRINCIPAL_TABLE, SERVER_PERMISSIONS_TABLE, SERVER_PRINCIPAL_SQL_LOGIN, SERVER_PRINCIPAL_WINDOWS_LOGIN, SERVER_PRINCIPAL_WINDOWS_GROUP)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(src)

		serverRoles, roleSrc, err := c.getPrincipals(SERVER_PRINCIPAL_TABLE, SERVER_PERMISSIONS_TABLE, SERVER_PRINCIPAL_SERVER_ROLE)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(roleSrc)

		members, memberSrc, err := c.getRoleMembers(SERVER_ROLE_MEMBER_TABLE)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(memberSrc)

		etlUsers, err := getEtlUsersAndRolesFromMssqlPrincipalsAndRoles(serverLogins, serverRoles, members)
		if err != nil {
			return nil, nil, err
		}

		for k, v := range etlUsers {
			sidToUser[k] = v
		}
	}

	{
		databaseLogins, src, err := c.getPrincipals(DATABASE_PRINCIPAL_TABLE, DATABASE_PERMISSIONS_TABLE, DATABASE_PRINCIPAL_EXTERNAL_USER_FROM_AD, DATABASE_PRINCIPAL_SQL_USER, DATABASE_PRINCIPAL_WINDOWS_USER, DATABASE_PRINCIPAL_WINDOWS_GROUP, DATABASE_PRINCIPAL_EXTERNAL_GROUP_FROM_AD)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(src)

		databaseRoles, roleSrc, err := c.getPrincipals(DATABASE_PRINCIPAL_TABLE, DATABASE_PERMISSIONS_TABLE, DATABASE_PRINCIPAL_DATABASE_ROLE, DATABASE_PRINCIPAL_APPLICATION_ROLE)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(roleSrc)

		members, memberSrc, err := c.getRoleMembers(DATABASE_ROLE_MEMBER_TABLE)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(memberSrc)

		etlUsers, err := getEtlUsersAndRolesFromMssqlPrincipalsAndRoles(databaseLogins, databaseRoles, members)
		if err != nil {
			return nil, nil, err
		}

		for k, v := range etlUsers {
			sidToUser[k].NestedUsers[v.Username] = v
		}
	}

	retUsers := []*types.EtlUser{}
	for _, u := range sidToUser {
		retUsers = append(retUsers, u)
	}

	return retUsers, finalSource, nil
}
