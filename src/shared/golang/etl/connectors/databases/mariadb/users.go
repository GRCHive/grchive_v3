package mariadb

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
)

type EtlMariadbConnectorUser struct {
	db *databases.DB
}

func (c *EtlMariadbConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	source := connectors.CreateSourceInfo()

	// Returns all users and roles as a user because there's nothing to really
	// distinguish between a "User" and a "Role". If the role is actively used, however, then the
	// role will show up as an EtlRole in the roles map for a particular EtlUser.
	allUsers := map[string]*types.EtlUser{}

	// First get all users + roles and their associated grants.
	{
		// Don't use a CTE here since < MariaDB 10.2 doesn't support it.
		rows, cmd, err := c.db.LoggedQuery(`
			SELECT 
				ur.Host,
				ur.User,
				p.object,
				p.permission
			FROM mysql.user AS ur
			LEFT JOIN (
				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Select' AS permission
				FROM mysql.user AS up
				WHERE up.Select_priv = 'Y'

				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Insert' AS permission
				FROM mysql.user AS up
				WHERE up.Insert_priv = 'Y'

				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Update' AS permission
				FROM mysql.user AS up
				WHERE up.Update_priv = 'Y'

				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Delete' AS permission
				FROM mysql.user AS up
				WHERE up.Delete_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Index' AS permission
				FROM mysql.user AS up
				WHERE up.Index_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Alter' AS permission
				FROM mysql.user AS up
				WHERE up.Alter_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Create' AS permission
				FROM mysql.user AS up
				WHERE up.Create_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Drop' AS permission
				FROM mysql.user AS up
				WHERE up.Drop_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Grant' AS permission
				FROM mysql.user AS up
				WHERE up.Grant_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Create_view' AS permission
				FROM mysql.user AS up
				WHERE up.Create_view_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Show_view' AS permission
				FROM mysql.user AS up
				WHERE up.Show_view_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Create_routine' AS permission
				FROM mysql.user AS up
				WHERE up.Create_routine_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Alter_routine' AS permission
				FROM mysql.user AS up
				WHERE up.Alter_routine_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Execute' AS permission
				FROM mysql.user AS up
				WHERE up.Execute_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Trigger' AS permission
				FROM mysql.user AS up
				WHERE up.Trigger_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Event' AS permission
				FROM mysql.user AS up
				WHERE up.Event_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Create_tmp_table' AS permission
				FROM mysql.user AS up
				WHERE up.Create_tmp_table_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Lock_tables' AS permission
				FROM mysql.user AS up
				WHERE up.Lock_tables_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'References' AS permission
				FROM mysql.user AS up
				WHERE up.References_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Reload' AS permission
				FROM mysql.user AS up
				WHERE up.Reload_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Shutdown' AS permission
				FROM mysql.user AS up
				WHERE up.Shutdown_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Process' AS permission
				FROM mysql.user AS up
				WHERE up.Process_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'File' AS permission
				FROM mysql.user AS up
				WHERE up.File_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Show_db' AS permission
				FROM mysql.user AS up
				WHERE up.Show_db_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Super' AS permission
				FROM mysql.user AS up
				WHERE up.Super_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Repl_slave' AS permission
				FROM mysql.user AS up
				WHERE up.Repl_slave_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Repl_client' AS permission
				FROM mysql.user AS up
				WHERE up.Repl_client_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Create_user' AS permission
				FROM mysql.user AS up
				WHERE up.Create_user_priv = 'Y'
				
				UNION

				SELECT
					up.Host AS host,
					up.User AS user,
					'User' AS object,
					'Create_tablespace' AS permission
				FROM mysql.user AS up
				WHERE up.Create_tablespace_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Select' AS permission
				FROM mysql.db AS db
				WHERE db.Select_priv = 'Y'

				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Insert' AS permission
				FROM mysql.db AS db
				WHERE db.Insert_priv = 'Y'

				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Update' AS permission
				FROM mysql.db AS db
				WHERE db.Update_priv = 'Y'

				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Delete' AS permission
				FROM mysql.db AS db
				WHERE db.Delete_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Index' AS permission
				FROM mysql.db AS db
				WHERE db.Index_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Alter' AS permission
				FROM mysql.db AS db
				WHERE db.Alter_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Create' AS permission
				FROM mysql.db AS db
				WHERE db.Create_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Drop' AS permission
				FROM mysql.db AS db
				WHERE db.Drop_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Grant' AS permission
				FROM mysql.db AS db
				WHERE db.Grant_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Create_view' AS permission
				FROM mysql.db AS db
				WHERE db.Create_view_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Show_view' AS permission
				FROM mysql.db AS db
				WHERE db.Show_view_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Create_routine' AS permission
				FROM mysql.db AS db
				WHERE db.Create_routine_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Alter_routine' AS permission
				FROM mysql.db AS db
				WHERE db.Alter_routine_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Execute' AS permission
				FROM mysql.db AS db
				WHERE db.Execute_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Trigger' AS permission
				FROM mysql.db AS db
				WHERE db.Trigger_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Event' AS permission
				FROM mysql.db AS db
				WHERE db.Event_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Create_tmp_table' AS permission
				FROM mysql.db AS db
				WHERE db.Create_tmp_table_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'Lock_tables' AS permission
				FROM mysql.db AS db
				WHERE db.Lock_tables_priv = 'Y'
				
				UNION

				SELECT
					db.Host AS host,
					db.User AS user,
					CONCAT('DB::', db.Db) AS object,
					'References' AS permission
				FROM mysql.db AS db
				WHERE db.References_priv = 'Y'

				UNION

				SELECT
					tp.Host AS host,
					tp.User AS user,
					CONCAT('TBL::', tp.Db, '.', tp.Table_name) AS object,
					tp.Table_priv AS permission
				FROM mysql.tables_priv AS tp
				WHERE tp.Table_priv IS NOT NULL AND tp.Table_priv != ''

				UNION

				SELECT
					tp.Host AS host,
					tp.User AS user,
					CONCAT('TBLCOL::', tp.Db, '.', tp.Table_name) AS object,
					tp.Column_priv AS permission
				FROM mysql.tables_priv AS tp
				WHERE tp.Column_priv IS NOT NULL AND tp.Column_priv != ''

				UNION

				SELECT
					cp.Host AS host,
					cp.User AS user,
					CONCAT('COL::', cp.Db, '.', cp.Table_name, '.', cp.Column_name) AS object,
					cp.Column_priv AS permission
				FROM mysql.columns_priv AS cp

				UNION

				SELECT
					pp.Host AS host,
					pp.User AS user,
					CONCAT(pp.Routine_type, '::', pp.Db, '.', pp.Routine_name) AS object,
					pp.Proc_priv AS permission
				FROM mysql.procs_priv AS pp
			) AS p
				ON p.host = ur.Host AND p.user = ur.User
		`)

		if err != nil {
			return nil, nil, err
		}
		source.AddCommand(cmd)
		defer rows.Close()

		for rows.Next() {
			type Result struct {
				Host       string  `db:"Host"`
				User       string  `db:"User"`
				Object     *string `db:"object"`
				Permission *string `db:"permission"`
			}

			result := Result{}
			err = rows.StructScan(&result)

			if err != nil {
				return nil, nil, err
			}

			// This is to maintain a parallel with what'd we expect the MySQL 8 database to give us.
			if result.Host == "" {
				result.Host = "%"
			}

			username := fmt.Sprintf("%s@%s", result.User, result.Host)
			user, ok := allUsers[username]
			if !ok {
				user = &types.EtlUser{
					Username: username,
					Roles: map[string]*types.EtlRole{
						"Self": &types.EtlRole{
							Name:        "Self",
							Permissions: map[string][]string{},
						},
					},
				}
			}

			if result.Object != nil && result.Permission != nil {
				role := user.Roles["Self"]
				perms, ok := role.Permissions[*result.Object]
				if !ok {
					perms = []string{}
				}
				perms = append(perms, *result.Permission)

				role.Permissions[*result.Object] = perms
			}
			allUsers[username] = user
		}
	}

	// Next get the relationships between users and roles.
	{
		rows, cmd, err := c.db.LoggedQuery(`
			SELECT Host, User, Role FROM mysql.roles_mapping;
		`)

		if err != nil {
			return nil, nil, err
		}
		source.AddCommand(cmd)
		defer rows.Close()

		for rows.Next() {
			type Result struct {
				Host string `db:"Host"`
				User string `db:"User"`
				Role string `db:"Role"`
			}

			result := Result{}
			err = rows.StructScan(&result)

			if err != nil {
				return nil, nil, err
			}

			toUserName := fmt.Sprintf("%s@%s", result.User, result.Host)
			fromRoleName := result.Role + "@%"

			user := allUsers[toUserName]
			role := allUsers[fromRoleName]
			user.Roles[fromRoleName] = &types.EtlRole{
				Name:        fromRoleName,
				Permissions: role.Roles["Self"].Permissions,
			}
		}
	}

	retUsers := []*types.EtlUser{}
	for _, v := range allUsers {
		retUsers = append(retUsers, v)
	}

	return retUsers, source, nil
}
