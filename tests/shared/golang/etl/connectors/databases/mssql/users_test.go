// +build !unit

package mssql

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/databases/mssql_utility"
	"testing"
	"time"
)

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, version := range []string{
		"2019-CU6-ubuntu-16.04",
		"2017-CU21-ubuntu-16.04",
	} {
		ctx := context.Background()
		expectedUsers := map[string]*types.EtlUser{}

		// Test cases:
		// 1) Create a server login - should have guest access.
		// 2) Create a server login with an associated database user
		// 3) Grant test permissions to a server login (server level)
		// 4) Grant test permissions to a database user (database level)
		// 5) Grant test permissions to a database user (table level)
		// 6) Role inheritance for database users
		// 7) Role inheritance for server logins
		container, db := mssql_utility.SetupMssqlDatabase(t, version, func(db *sqlx.DB) error {
			tx := db.MustBegin()

			gstNow := time.Date(2003, 4, 8, 9, 10, 19, 647000000, time.UTC)
			guestUser := &types.EtlUser{
				Username:    "guest",
				CreatedTime: &gstNow,
				Roles: map[string]*types.EtlRole{
					"guest": &types.EtlRole{
						Name: "guest",
						Permissions: types.PermissionMap{
							"DATABASE::::.": []string{"CONNECT"},
						},
						Denied: types.PermissionMap{},
					},
				},
				NestedUsers: map[string]*types.EtlUser{},
			}

			// Test 1
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_1 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_1"] = &types.EtlUser{
					Username: "test_1",
					Roles: map[string]*types.EtlRole{
						"test_1": &types.EtlRole{
							Name: "test_1",
							Permissions: map[string][]string{
								"SERVER::::.": []string{"CONNECT SQL"},
							},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
					},
				}
			}

			// Test 2
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_2 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE USER test_2_user FOR LOGIN test_2`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_2"] = &types.EtlUser{
					Username: "test_2",
					Roles: map[string]*types.EtlRole{
						"test_2": &types.EtlRole{
							Name: "test_2",
							Permissions: map[string][]string{
								"SERVER::::.": []string{"CONNECT SQL"},
							},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
						"test_2_user": &types.EtlUser{
							Username:    "test_2_user",
							CreatedTime: &nw,
							Roles: map[string]*types.EtlRole{
								"test_2_user": &types.EtlRole{
									Name: "test_2_user",
									Permissions: types.PermissionMap{
										"DATABASE::::.": []string{"CONNECT"},
									},
									Denied: types.PermissionMap{},
								},
							},
							NestedUsers: map[string]*types.EtlUser{},
						},
					},
				}
			}

			// Test 3
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_3 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT ALTER SETTINGS TO test_3`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_3"] = &types.EtlUser{
					Username: "test_3",
					Roles: map[string]*types.EtlRole{
						"test_3": &types.EtlRole{
							Name: "test_3",
							Permissions: map[string][]string{
								"SERVER::::.": []string{"CONNECT SQL", "ALTER SETTINGS"},
							},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
					},
				}
			}

			// Test 4
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_4 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE USER test_4_user FOR LOGIN test_4`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT CREATE TABLE TO test_4_user`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_4"] = &types.EtlUser{
					Username: "test_4",
					Roles: map[string]*types.EtlRole{
						"test_4": &types.EtlRole{
							Name: "test_4",
							Permissions: types.PermissionMap{
								"SERVER::::.": []string{"CONNECT SQL"},
							},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
						"test_4_user": &types.EtlUser{
							Username:    "test_4_user",
							CreatedTime: &nw,
							Roles: map[string]*types.EtlRole{
								"test_4_user": &types.EtlRole{
									Name: "test_4_user",
									Permissions: types.PermissionMap{
										"DATABASE::::.": []string{"CONNECT", "CREATE TABLE"},
									},
									Denied: types.PermissionMap{},
								},
							},
							NestedUsers: map[string]*types.EtlUser{},
						},
					},
				}
			}

			// Test 5
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_5 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE USER test_5_user FOR LOGIN test_5`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE TABLE dbo.Test ( Id INT PRIMARY KEY )`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`GRANT SELECT ON OBJECT::dbo.Test TO test_5_user`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_5"] = &types.EtlUser{
					Username: "test_5",
					Roles: map[string]*types.EtlRole{
						"test_5": &types.EtlRole{
							Name: "test_5",
							Permissions: types.PermissionMap{
								"SERVER::::.": []string{"CONNECT SQL"},
							},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
						"test_5_user": &types.EtlUser{
							Username:    "test_5_user",
							CreatedTime: &nw,
							Roles: map[string]*types.EtlRole{
								"test_5_user": &types.EtlRole{
									Name: "test_5_user",
									Permissions: types.PermissionMap{
										"DATABASE::::.":                          []string{"CONNECT"},
										"OBJECT_OR_COLUMN::USER_TABLE::dbo.Test": []string{"SELECT"},
									},
									Denied: types.PermissionMap{},
								},
							},
							NestedUsers: map[string]*types.EtlUser{},
						},
					},
				}
			}

			// Test 6
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_6 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE USER test_6_user FOR LOGIN test_6`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE test_6_role1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE ROLE test_6_role2`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`ALTER ROLE test_6_role1 ADD MEMBER test_6_user`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`ALTER ROLE test_6_role2 ADD MEMBER test_6_role1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_6"] = &types.EtlUser{
					Username: "test_6",
					Roles: map[string]*types.EtlRole{
						"test_6": &types.EtlRole{
							Name: "test_6",
							Permissions: map[string][]string{
								"SERVER::::.": []string{"CONNECT SQL"},
							},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
						"test_6_user": &types.EtlUser{
							Username:    "test_6_user",
							CreatedTime: &nw,
							Roles: map[string]*types.EtlRole{
								"test_6_user": &types.EtlRole{
									Name: "test_6_user",
									Permissions: types.PermissionMap{
										"DATABASE::::.": []string{"CONNECT"},
									},
									Denied: types.PermissionMap{},
								},
								"test_6_role1": &types.EtlRole{
									Name:        "test_6_role1",
									Permissions: map[string][]string{},
								},
								"test_6_role2": &types.EtlRole{
									Name:        "test_6_role2",
									Permissions: map[string][]string{},
								},
							},
							NestedUsers: map[string]*types.EtlUser{},
						},
					},
				}
			}

			// Test 7
			{
				nw := time.Now()
				_, err := tx.Exec(`CREATE LOGIN test_7 WITH PASSWORD = 'qpwoeiru0A!'`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE SERVER ROLE test_7_role1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`CREATE SERVER ROLE test_7_role2`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`ALTER SERVER ROLE test_7_role1 ADD MEMBER test_7`)
				if err != nil {
					tx.Rollback()
					return err
				}

				_, err = tx.Exec(`ALTER SERVER ROLE test_7_role2 ADD MEMBER test_7_role1`)
				if err != nil {
					tx.Rollback()
					return err
				}

				expectedUsers["test_7"] = &types.EtlUser{
					Username: "test_7",
					Roles: map[string]*types.EtlRole{
						"test_7": &types.EtlRole{
							Name: "test_7",
							Permissions: map[string][]string{
								"SERVER::::.": []string{"CONNECT SQL"},
							},
						},
						"test_7_role1": &types.EtlRole{
							Name:        "test_7_role1",
							Permissions: map[string][]string{},
						},
						"test_7_role2": &types.EtlRole{
							Name:        "test_7_role2",
							Permissions: map[string][]string{},
						},
					},
					CreatedTime: &nw,
					NestedUsers: map[string]*types.EtlUser{
						"guest": guestUser,
					},
				}
			}

			return tx.Commit()
		})

		defer container.Terminate(ctx)

		connector, err := createMssqlConnectorUser(&databases.DB{SqlxLike: db})
		if err != nil {
			t.Error(err)
		}

		users, _, err := connector.GetUserListing()
		if err != nil {
			t.Error(err)
		}
		test_utility.CompareUserListing(g, users, expectedUsers, test_utility.CompareUserListingOptions{
			UsersToIgnore: []string{
				"sa",
				"BUILTIN\\Administrators",
				"##MS_PolicyEventProcessingLogin##",
				"##MS_PolicyTsqlExecutionLogin##",
			},
			PermissionObjectsToIgnore: []string{},
		})

	}
}
