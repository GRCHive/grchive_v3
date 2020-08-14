package psql

import (
	"database/sql"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
)

const PsqlSuperPermission = "rolsuper"
const PsqlCreateRolePermission = "rolcreaterole"
const PsqlCreateDbPermission = "rolcreatedb"
const PsqlReplicationPermission = "rolreplication"

type EtlPsqlConnectorUser struct {
	db *databases.DB
}

func createPsqlConnectorUser(db *databases.DB) (*EtlPsqlConnectorUser, error) {
	return &EtlPsqlConnectorUser{
		db: db,
	}, nil
}

// Obtain users and what roles/permissions they have.
// We define a "user" as a role that can login.
// Permissions in this case would be grants. We also need to also account for the boolean permissions of
// rolsuper/rolcreaterole/rolcreatedb/rolbypassrls/rolreplication.
func (c *EtlPsqlConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	// Get all users, roles they're a part of, and grants they have all in one SQL query.
	source := connectors.CreateSourceInfo()
	rows, cmd, err := c.db.LoggedQuery(`
		WITH RECURSIVE parents AS (
			SELECT
				u.rolname as child_role,
				pr.rolname as parent_role
			FROM users AS u
			INNER JOIN pg_auth_members AS pam
				ON pam.member = u.oid
			INNER JOIN pg_roles AS pr
				ON pr.oid = pam.roleid

			UNION

			SELECT
				p.child_role as child_role,
				ppr.rolname as parent_role
			FROM pg_roles AS ppr
			INNER JOIN pg_auth_members AS pam
				ON pam.roleid = ppr.oid
			INNER JOIN pg_roles AS pr
				ON pr.oid = pam.member
			INNER JOIN parents AS p
				ON p.parent_role = pr.rolname
		), permissions AS (
			SELECT
				tp.grantee,
				CASE WHEN tp.grantee IS NULL THEN ''
					 ELSE CONCAT('TBL::', tp.table_catalog, '.', tp.table_schema, '.', tp.table_name)
				END AS object,
				COALESCE(tp.privilege_type, '') as permission
			FROM information_schema.table_privileges AS tp

			UNION

			SELECT
				cp.grantee,
				CASE WHEN cp.grantee IS NULL THEN ''
					 ELSE CONCAT('COLUMN::', cp.table_catalog, '.', cp.table_schema, '.', cp.table_name, '.', cp.column_name)
				END AS object,
				COALESCE(cp.privilege_type, '') as permission
			FROM information_schema.column_privileges AS cp

			UNION

			SELECT
				rp.grantee,
				CASE WHEN rp.grantee IS NULL THEN ''
					 ELSE CONCAT('ROUTINE::', rp.specific_catalog, '.', rp.specific_schema, '.', rp.specific_name)
				END AS object,
				COALESCE(rp.privilege_type, '') as permission
			FROM information_schema.routine_privileges AS rp

			UNION

			SELECT
				up.grantee,
				CASE WHEN up.grantee IS NULL THEN ''
					 ELSE CONCAT(up.object_type, '::', up.object_catalog, '.', up.object_schema, '.', up.object_name)
				END AS object,
				COALESCE(up.privilege_type, '') as permission
			FROM information_schema.usage_privileges AS up
		), users AS (
			SELECT
				pr.oid,
				pr.rolname,
				pr.rolsuper,
				pr.rolcreaterole,
				pr.rolcreatedb,
				pr.rolreplication
			FROM pg_roles AS pr
			WHERE pr.rolcanlogin = true
		) 
		SELECT
			u.rolname AS "Username",
			u.rolsuper AS "Super",
			u.rolcreaterole AS "CreateRole",
			u.rolcreatedb AS "CreateDb",
			u.rolreplication AS "Replication",
			'Self' AS "ParentRole",
			p.object AS "Object",
			p.permission AS "Permission"
		FROM users AS u			
		LEFT JOIN permissions AS p
			ON p.grantee = u.rolname

		UNION
		SELECT
			u.rolname AS "Username",
			u.rolsuper AS "Super",
			u.rolcreaterole AS "CreateRole",
			u.rolcreatedb AS "CreateDb",
			u.rolreplication AS "Replication",
			par.parent_role AS "ParentRole",
			p.object AS "Object",
			p.permission AS "Permission"
		FROM users AS u			
		INNER JOIN parents AS par
			ON par.child_role = u.rolname
		LEFT JOIN permissions AS p
			ON p.grantee = par.parent_role
	`)

	if err != nil {
		return nil, nil, err
	}
	source.AddCommand(cmd)

	// Need to keep a map of all users see that we can easily
	// aggregate all the permissions into a single user object.
	userMap := map[string]*types.EtlUser{}
	allUsers := []*types.EtlUser{}

	defer rows.Close()
	for rows.Next() {
		type Result struct {
			Username    string         `db:"Username"`
			Super       bool           `db:"Super"`
			CreateRole  bool           `db:"CreateRole"`
			CreateDb    bool           `db:"CreateDb"`
			Replication bool           `db:"Replication"`
			ParentRole  string         `db:"ParentRole"`
			Object      sql.NullString `db:"Object"`
			Permission  sql.NullString `db:"Permission"`
		}
		result := Result{}
		err = rows.StructScan(&result)
		if err != nil {
			return nil, nil, err
		}

		user, ok := userMap[result.Username]
		if !ok {
			user = &types.EtlUser{
				Username: result.Username,
				Roles:    map[string]*types.EtlRole{},
			}
			userMap[result.Username] = user
			allUsers = append(allUsers, user)

			// Only handle the Super/CreateRole/CreateDb/Replication
			// permissions once as they'll be present in every row.
			permissions := []string{}

			if result.Super {
				permissions = append(permissions, PsqlSuperPermission)
			}

			if result.CreateRole {
				permissions = append(permissions, PsqlCreateRolePermission)
			}

			if result.CreateDb {
				permissions = append(permissions, PsqlCreateDbPermission)
			}

			if result.Replication {
				permissions = append(permissions, PsqlReplicationPermission)
			}

			user.Roles["Self"] = &types.EtlRole{
				Name: "Self",
				Permissions: map[string][]string{
					"Self": permissions,
				},
			}
		}

		role, ok := user.Roles[result.ParentRole]
		if !ok {
			role = &types.EtlRole{
				Name:        result.ParentRole,
				Permissions: map[string][]string{},
			}
			user.Roles[result.ParentRole] = role
		}

		if result.Object.Valid && result.Permission.Valid {
			permissions, ok := role.Permissions[result.Object.String]
			if !ok {
				permissions = []string{}
				role.Permissions[result.Object.String] = permissions
			}

			role.Permissions[result.Object.String] = append(permissions, result.Permission.String)
		}
	}

	return allUsers, source, nil
}
