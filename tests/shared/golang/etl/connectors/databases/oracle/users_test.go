// +build !unit

package oracle

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/databases/oracle_utility"
	"testing"
	"time"
)

var v19cUsers = []string{
	"SYS",
	"SYSTEM",
	"XS$NULL",
	"DBSNMP",
	"APPQOSSYS",
	"GSMCATUSER",
	"MDDATA",
	"DBSFWUSER",
	"SYSBACKUP",
	"REMOTE_SCHEDULER_AGENT",
	"GGSYS",
	"ANONYMOUS",
	"GSMUSER",
	"SYSRAC",
	"GSMROOTUSER",
	"CTXSYS",
	"OJVMSYS",
	"DVSYS",
	"DVF",
	"SI_INFORMTN_SCHEMA",
	"AUDSYS",
	"GSMADMIN_INTERNAL",
	"DIP",
	"ORDPLUGINS",
	"LBACSYS",
	"MDSYS",
	"OLAPSYS",
	"SYSKM",
	"ORDDATA",
	"OUTLN",
	"SYS$UMF",
	"ORACLE_OCM",
	"XDB",
	"WMSYS",
	"SYSDG",
	"ORDSYS",
}

var v18cUsers = []string{
	"SYS",
	"SYSTEM",
	"XS$NULL",
	"OJVMSYS",
	"LBACSYS",
	"OUTLN",
	"SYS$UMF",
	"DBSNMP",
	"APPQOSSYS",
	"DBSFWUSER",
	"GGSYS",
	"ANONYMOUS",
	"CTXSYS",
	"DVF",
	"SI_INFORMTN_SCHEMA",
	"DVSYS",
	"GSMADMIN_INTERNAL",
	"ORDPLUGINS",
	"MDSYS",
	"OLAPSYS",
	"ORDDATA",
	"XDB",
	"WMSYS",
	"ORDSYS",
	"GSMCATUSER",
	"MDDATA",
	"SYSBACKUP",
	"REMOTE_SCHEDULER_AGENT",
	"GSMUSER",
	"SYSRAC",
	"AUDSYS",
	"DIP",
	"SYSKM",
	"ORACLE_OCM",
	"SYSDG",
}

var v12cUsers = []string{
	"SYS",
	"SYSTEM",
	"XS$NULL",
	"OJVMSYS",
	"LBACSYS",
	"OUTLN",
	"SYS$UMF",
	"DBSNMP",
	"APPQOSSYS",
	"DBSFWUSER",
	"GGSYS",
	"ANONYMOUS",
	"CTXSYS",
	"SI_INFORMTN_SCHEMA",
	"DVSYS",
	"DVF",
	"GSMADMIN_INTERNAL",
	"ORDPLUGINS",
	"MDSYS",
	"OLAPSYS",
	"ORDDATA",
	"XDB",
	"WMSYS",
	"ORDSYS",
	"GSMCATUSER",
	"MDDATA",
	"SYSBACKUP",
	"REMOTE_SCHEDULER_AGENT",
	"GSMUSER",
	"SYSRAC",
	"AUDSYS",
	"DIP",
	"SYSKM",
	"ORACLE_OCM",
	"SYSDG",
	"SPATIAL_CSW_ADMIN_USR",
}

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Version     string
		SystemUsers []string
	}{
		{
			Version:     "19.3.0-ee",
			SystemUsers: v19cUsers,
		},
		{
			Version:     "18.3.0-ee",
			SystemUsers: v18cUsers,
		},
		{
			Version:     "12.2.0.1-ee",
			SystemUsers: v12cUsers,
		},
	} {
		ctx := context.Background()
		expectedUsers := map[string]*types.EtlUser{}

		// Test cases:
		// 1) Create user and ensure it is returned
		// 2) Create user, assign to a role
		// 3) Nested role: create user, assign to a role, assign a role to that role
		// 4) User/Role: System privileges
		// 5) User/Role: Table privileges
		// 6) User/Role: Column privileges
		container, db := oracle_utility.SetupOracleDatabase(t, test.Version, func(db *sqlx.DB) error {
			tx := db.MustBegin()

			{
				now := time.Now()
				_, err := tx.Exec(`CREATE USER C##TEST1 IDENTIFIED BY password1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["C##TEST1"] = &types.EtlUser{
					Username:    "C##TEST1",
					CreatedTime: &now,
					Roles: map[string]*types.EtlRole{
						"C##TEST1": &types.EtlRole{
							Name:        "C##TEST1",
							Permissions: types.PermissionMap{},
						},
					},
				}
			}

			{
				now := time.Now()
				_, err := tx.Exec(`CREATE USER C##TEST2 IDENTIFIED BY password1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE C##TESTROLE1 NOT IDENTIFIED`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT C##TESTROLE1 TO C##TEST2`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["C##TEST2"] = &types.EtlUser{
					Username:    "C##TEST2",
					CreatedTime: &now,
					Roles: map[string]*types.EtlRole{
						"C##TEST2": &types.EtlRole{
							Name:        "C##TEST2",
							Permissions: types.PermissionMap{},
						},
						"C##TESTROLE1": &types.EtlRole{
							Name:        "C##TESTROLE1",
							Permissions: types.PermissionMap{},
						},
					},
				}
			}

			{
				now := time.Now()
				_, err := tx.Exec(`CREATE USER C##TEST3 IDENTIFIED BY password1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE C##TESTROLE2 NOT IDENTIFIED`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE C##TESTROLE3 NOT IDENTIFIED`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT C##TESTROLE2 TO C##TEST3`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT C##TESTROLE3 TO C##TESTROLE2`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["C##TEST3"] = &types.EtlUser{
					Username:    "C##TEST3",
					CreatedTime: &now,
					Roles: map[string]*types.EtlRole{
						"C##TEST3": &types.EtlRole{
							Name:        "C##TEST3",
							Permissions: types.PermissionMap{},
						},
						"C##TESTROLE2": &types.EtlRole{
							Name:        "C##TESTROLE2",
							Permissions: types.PermissionMap{},
						},
						"C##TESTROLE3": &types.EtlRole{
							Name:        "C##TESTROLE3",
							Permissions: types.PermissionMap{},
						},
					},
				}
			}

			{
				now := time.Now()
				_, err := tx.Exec(`CREATE USER C##TEST4 IDENTIFIED BY password1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE C##TESTROLE4 NOT IDENTIFIED`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT C##TESTROLE4 TO C##TEST4`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT ALTER ANY TABLE TO C##TEST4`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT ALTER ANY INDEX TO C##TESTROLE4`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["C##TEST4"] = &types.EtlUser{
					Username:    "C##TEST4",
					CreatedTime: &now,
					Roles: map[string]*types.EtlRole{
						"C##TEST4": &types.EtlRole{
							Name: "C##TEST4",
							Permissions: types.PermissionMap{
								"SYSTEM": []string{"ALTER ANY TABLE"},
							},
						},
						"C##TESTROLE4": &types.EtlRole{
							Name: "C##TESTROLE4",
							Permissions: types.PermissionMap{
								"SYSTEM": []string{"ALTER ANY INDEX"},
							},
						},
					},
				}
			}

			{
				now := time.Now()
				_, err := tx.Exec(`CREATE USER C##TEST5 IDENTIFIED BY password1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE C##TESTROLE5 NOT IDENTIFIED`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT C##TESTROLE5 TO C##TEST5`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE TABLE TESTTABLE1 ( id NUMBER GENERATED BY DEFAULT AS IDENTITY )`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT SELECT ON TESTTABLE1 TO C##TEST5`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT DELETE ON TESTTABLE1 TO C##TESTROLE5`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["C##TEST5"] = &types.EtlUser{
					Username:    "C##TEST5",
					CreatedTime: &now,
					Roles: map[string]*types.EtlRole{
						"C##TEST5": &types.EtlRole{
							Name: "C##TEST5",
							Permissions: types.PermissionMap{
								"TABLE::TESTTABLE1": []string{"SELECT"},
							},
						},
						"C##TESTROLE5": &types.EtlRole{
							Name: "C##TESTROLE5",
							Permissions: types.PermissionMap{
								"TABLE::TESTTABLE1": []string{"DELETE"},
							},
						},
					},
				}
			}

			{
				now := time.Now()
				_, err := tx.Exec(`CREATE USER C##TEST6 IDENTIFIED BY password1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE C##TESTROLE6 NOT IDENTIFIED`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT C##TESTROLE6 TO C##TEST6`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE TABLE TESTTABLE2 ( id NUMBER GENERATED BY DEFAULT AS IDENTITY )`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT INSERT (id) ON TESTTABLE2 TO C##TEST6`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT UPDATE (id) ON TESTTABLE2 TO C##TESTROLE6`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["C##TEST6"] = &types.EtlUser{
					Username:    "C##TEST6",
					CreatedTime: &now,
					Roles: map[string]*types.EtlRole{
						"C##TEST6": &types.EtlRole{
							Name: "C##TEST6",
							Permissions: types.PermissionMap{
								"COLUMN::TESTTABLE2.ID": []string{"INSERT"},
							},
						},
						"C##TESTROLE6": &types.EtlRole{
							Name: "C##TESTROLE6",
							Permissions: types.PermissionMap{
								"COLUMN::TESTTABLE2.ID": []string{"UPDATE"},
							},
						},
					},
				}
			}

			return tx.Commit()
		})

		defer container.Terminate(ctx)

		connector, err := createOracleConnectorUser(&databases.DB{SqlxLike: db})
		if err != nil {
			t.Error(err)
		}

		users, _, err := connector.GetUserListing()
		if err != nil {
			t.Error(err)
		}
		test_utility.CompareUserListing(g, users, expectedUsers, test_utility.CompareUserListingOptions{
			UsersToIgnore:             test.SystemUsers,
			PermissionObjectsToIgnore: []string{},
		})
	}
}
