// +build !unit
package v5

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/databases/mysql_utility"
	"strings"
	"testing"
)

var v57UsersToIgnore = []string{"root@%", "root@localhost", "mysql.sys@localhost", "mysql.session@localhost"}
var v56UsersToIgnore = []string{"root@%", "root@localhost"}

var mysqlGlobalPrivileges = []struct {
	Grant      string
	Permission string
}{
	{
		Grant:      "ALTER",
		Permission: "Alter",
	},
	{
		Grant:      "ALTER ROUTINE",
		Permission: "Alter_routine",
	},
	{
		Grant:      "CREATE",
		Permission: "Create",
	},
	{
		Grant:      "CREATE ROUTINE",
		Permission: "Create_routine",
	},
	{
		Grant:      "CREATE TABLESPACE",
		Permission: "Create_tablespace",
	},
	{
		Grant:      "CREATE TEMPORARY TABLES",
		Permission: "Create_tmp_table",
	},
	{
		Grant:      "CREATE USER",
		Permission: "Create_user",
	},
	{
		Grant:      "CREATE VIEW",
		Permission: "Create_view",
	},
	{
		Grant:      "DELETE",
		Permission: "Delete",
	},
	{
		Grant:      "DROP",
		Permission: "Drop",
	},
	{
		Grant:      "EVENT",
		Permission: "Event",
	},
	{
		Grant:      "EXECUTE",
		Permission: "Execute",
	},
	{
		Grant:      "FILE",
		Permission: "File",
	},
	{
		Grant:      "GRANT OPTION",
		Permission: "Grant",
	},
	{
		Grant:      "INDEX",
		Permission: "Index",
	},
	{
		Grant:      "INSERT",
		Permission: "Insert",
	},
	{
		Grant:      "LOCK TABLES",
		Permission: "Lock_tables",
	},
	{
		Grant:      "PROCESS",
		Permission: "Process",
	},
	{
		Grant:      "REFERENCES",
		Permission: "References",
	},
	{
		Grant:      "RELOAD",
		Permission: "Reload",
	},
	{
		Grant:      "REPLICATION CLIENT",
		Permission: "Repl_client",
	},
	{
		Grant:      "REPLICATION SLAVE",
		Permission: "Repl_slave",
	},
	{
		Grant:      "SELECT",
		Permission: "Select",
	},
	{
		Grant:      "SHOW DATABASES",
		Permission: "Show_db",
	},
	{
		Grant:      "SHOW VIEW",
		Permission: "Show_view",
	},
	{
		Grant:      "SHUTDOWN",
		Permission: "Shutdown",
	},
	{
		Grant:      "SUPER",
		Permission: "Super",
	},
	{
		Grant:      "TRIGGER",
		Permission: "Trigger",
	},
	{
		Grant:      "UPDATE",
		Permission: "Update",
	},
}

var mysqlDatabasePrivileges = []struct {
	Grant      string
	Permission string
}{
	{
		Grant:      "ALTER",
		Permission: "Alter",
	},
	{
		Grant:      "ALTER ROUTINE",
		Permission: "Alter_routine",
	},
	{
		Grant:      "CREATE",
		Permission: "Create",
	},
	{
		Grant:      "CREATE ROUTINE",
		Permission: "Create_routine",
	},
	{
		Grant:      "CREATE TEMPORARY TABLES",
		Permission: "Create_tmp_table",
	},
	{
		Grant:      "CREATE VIEW",
		Permission: "Create_view",
	},
	{
		Grant:      "DELETE",
		Permission: "Delete",
	},
	{
		Grant:      "DROP",
		Permission: "Drop",
	},
	{
		Grant:      "EVENT",
		Permission: "Event",
	},
	{
		Grant:      "EXECUTE",
		Permission: "Execute",
	},
	{
		Grant:      "GRANT OPTION",
		Permission: "Grant",
	},
	{
		Grant:      "INDEX",
		Permission: "Index",
	},
	{
		Grant:      "INSERT",
		Permission: "Insert",
	},
	{
		Grant:      "LOCK TABLES",
		Permission: "Lock_tables",
	},
	{
		Grant:      "REFERENCES",
		Permission: "References",
	},
	{
		Grant:      "SELECT",
		Permission: "Select",
	},
	{
		Grant:      "SHOW VIEW",
		Permission: "Show_view",
	},
	{
		Grant:      "TRIGGER",
		Permission: "Trigger",
	},
	{
		Grant:      "UPDATE",
		Permission: "Update",
	},
}

var mysqlTablePrivileges = []struct {
	Grant      string
	Permission string
}{
	{
		Grant:      "ALTER",
		Permission: "Alter",
	},
	{
		Grant:      "CREATE VIEW",
		Permission: "Create View",
	},
	{
		Grant:      "CREATE",
		Permission: "Create",
	},
	{
		Grant:      "DELETE",
		Permission: "Delete",
	},
	{
		Grant:      "DROP",
		Permission: "Drop",
	},
	{
		Grant:      "GRANT OPTION",
		Permission: "Grant",
	},
	{
		Grant:      "INDEX",
		Permission: "Index",
	},
	{
		Grant:      "INSERT",
		Permission: "Insert",
	},
	{
		Grant:      "REFERENCES",
		Permission: "References",
	},
	{
		Grant:      "SELECT",
		Permission: "Select",
	},
	{
		Grant:      "SHOW VIEW",
		Permission: "Show view",
	},
	{
		Grant:      "TRIGGER",
		Permission: "Trigger",
	},
	{
		Grant:      "UPDATE",
		Permission: "Update",
	},
}

var mysqlColumnPrivileges = []struct {
	Grant      string
	Permission string
}{
	{
		Grant:      "INSERT",
		Permission: "Insert",
	},
	{
		Grant:      "REFERENCES",
		Permission: "References",
	},
	{
		Grant:      "SELECT",
		Permission: "Select",
	},
	{
		Grant:      "UPDATE",
		Permission: "Update",
	},
}

var mysqlRoutinePrivileges = []struct {
	Grant      string
	Permission string
}{
	{
		Grant:      "ALTER ROUTINE",
		Permission: "Alter Routine",
	},
	{
		Grant:      "EXECUTE",
		Permission: "Execute",
	},
	{
		Grant:      "GRANT OPTION",
		Permission: "Grant",
	},
}

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, version := range []string{
		"5.7.31",
		"5.6.49",
	} {
		ctx := context.Background()

		refUsers := map[string]*types.EtlUser{}
		// Test cases:
		// 1) Test Global Privileges
		// 2) Test database level privileges
		// 3) Test table level privileges
		// 4) Test column level privileges
		// 5) Test process level privileges
		container, db := mysql_utility.SetupMySQLDatabase(t, version, func(db *sqlx.DB) error {
			tx := db.MustBegin()

			userCount := 0

			// Setup test users for testing global privileges.
			for _, priv := range mysqlGlobalPrivileges {
				username := fmt.Sprintf("test_user_%d", userCount)
				usernameWithHost := fmt.Sprintf("%s@%%", username)

				if _, err := tx.Exec(fmt.Sprintf(`
					CREATE USER '%s'
				`, username)); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(fmt.Sprintf(`
					GRANT %s ON *.* TO %s
				`, priv.Grant, username)); err != nil {
					tx.Rollback()
					return err
				}

				refUsers[usernameWithHost] = &types.EtlUser{
					Username: usernameWithHost,
					Roles: map[string]*types.EtlRole{
						"Self": &types.EtlRole{
							Name: "Self",
							Permissions: map[string][]string{
								"User": []string{priv.Permission},
							},
						},
					},
				}

				userCount += 1
			}

			// Setup test users for testing database privileges
			if _, err := tx.Exec(`CREATE DATABASE test_db`); err != nil {
				tx.Rollback()
				return err
			}

			for _, priv := range mysqlDatabasePrivileges {
				username := fmt.Sprintf("test_user_%d", userCount)
				usernameWithHost := fmt.Sprintf("%s@%%", username)

				if _, err := tx.Exec(fmt.Sprintf(`
					CREATE USER '%s'
				`, username)); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(fmt.Sprintf(`
					GRANT %s ON test_db.* TO %s
				`, priv.Grant, username)); err != nil {
					tx.Rollback()
					return err
				}

				refUsers[usernameWithHost] = &types.EtlUser{
					Username: usernameWithHost,
					Roles: map[string]*types.EtlRole{
						"Self": &types.EtlRole{
							Name: "Self",
							Permissions: map[string][]string{
								"DB::test_db": []string{priv.Permission},
							},
						},
					},
				}

				userCount += 1
			}

			// Setup test users for testing table privileges
			if _, err := tx.Exec(`CREATE DATABASE test_table`); err != nil {
				tx.Rollback()
				return err
			}

			if _, err := tx.Exec(`CREATE TABLE test_table.t1(
				id INTEGER PRIMARY KEY
			)`); err != nil {
				tx.Rollback()
				return err
			}

			for _, priv := range mysqlTablePrivileges {
				username := fmt.Sprintf("test_user_%d", userCount)
				usernameWithHost := fmt.Sprintf("%s@%%", username)

				if _, err := tx.Exec(fmt.Sprintf(`
					CREATE USER '%s'
				`, username)); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(fmt.Sprintf(`
					GRANT %s ON test_table.t1 TO %s
				`, priv.Grant, username)); err != nil {
					tx.Rollback()
					return err
				}

				refUsers[usernameWithHost] = &types.EtlUser{
					Username: usernameWithHost,
					Roles: map[string]*types.EtlRole{
						"Self": &types.EtlRole{
							Name: "Self",
							Permissions: map[string][]string{
								"TBL::test_table.t1": []string{priv.Permission},
							},
						},
					},
				}

				userCount += 1
			}

			// Setup test users for testing column privileges
			if _, err := tx.Exec(`CREATE DATABASE test_column`); err != nil {
				tx.Rollback()
				return err
			}

			if _, err := tx.Exec(`CREATE TABLE test_column.t1(
				id INTEGER PRIMARY KEY
			)`); err != nil {
				tx.Rollback()
				return err
			}

			for _, priv := range mysqlColumnPrivileges {
				username := fmt.Sprintf("test_user_%d", userCount)
				usernameWithHost := fmt.Sprintf("%s@%%", username)

				if _, err := tx.Exec(fmt.Sprintf(`
					CREATE USER '%s'
				`, username)); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(fmt.Sprintf(`
					GRANT %s (id) ON test_column.t1 TO %s
				`, priv.Grant, username)); err != nil {
					tx.Rollback()
					return err
				}

				refUsers[usernameWithHost] = &types.EtlUser{
					Username: usernameWithHost,
					Roles: map[string]*types.EtlRole{
						"Self": &types.EtlRole{
							Name: "Self",
							Permissions: map[string][]string{
								"TBLCOL::test_column.t1": []string{priv.Permission},
								"COL::test_column.t1.id": []string{priv.Permission},
							},
						},
					},
				}

				userCount += 1
			}

			// Setup test users for testing routine privileges
			if _, err := tx.Exec(`CREATE DATABASE test_routine`); err != nil {
				tx.Rollback()
				return err
			}

			if _, err := tx.Exec(`CREATE PROCEDURE test_routine.proc1() BEGIN END`); err != nil {
				tx.Rollback()
				return err
			}

			if _, err := tx.Exec(`
			CREATE FUNCTION test_routine.proc2()
			RETURNS INTEGER DETERMINISTIC

			BEGIN
				RETURN 0;
			END`); err != nil {
				tx.Rollback()
				return err
			}

			for _, priv := range mysqlRoutinePrivileges {
				username := fmt.Sprintf("test_user_%d", userCount)
				usernameWithHost := fmt.Sprintf("%s@%%", username)

				if _, err := tx.Exec(fmt.Sprintf(`
					CREATE USER '%s'
				`, username)); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(fmt.Sprintf(`
					GRANT %s ON PROCEDURE test_routine.proc1 TO %s
				`, priv.Grant, username)); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(fmt.Sprintf(`
					GRANT %s ON FUNCTION test_routine.proc2 TO %s
				`, priv.Grant, username)); err != nil {
					tx.Rollback()
					return err
				}

				refUsers[usernameWithHost] = &types.EtlUser{
					Username: usernameWithHost,
					Roles: map[string]*types.EtlRole{
						"Self": &types.EtlRole{
							Name: "Self",
							Permissions: map[string][]string{
								"PROCEDURE::test_routine.proc1": []string{priv.Permission},
								"FUNCTION::test_routine.proc2":  []string{priv.Permission},
							},
						},
					},
				}

				userCount += 1
			}

			return tx.Commit()
		})
		defer container.Terminate(ctx)

		connector := &EtlMysqlV5ConnectorUser{db: &databases.DB{
			SqlxLike: db,
		}}
		users, source, err := connector.GetUserListing()
		g.Expect(err).To(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))

		opts := test_utility.CompareUserListingOptions{
			PermissionObjectsToIgnore: []string{},
		}

		if strings.HasPrefix(version, "5.7") {
			opts.UsersToIgnore = v57UsersToIgnore
		} else if strings.HasPrefix(version, "5.6") {
			opts.UsersToIgnore = v56UsersToIgnore
		}

		test_utility.CompareUserListing(g, users, refUsers, opts)
	}
}
