// +build !unit

package psql

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/databases/psql_utility"
	"sort"
	"strings"
	"testing"
)

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, version := range []string{
		"12.3",
		"11.8",
		"10.13",
		"9.6.18",
		"9.5.22",
	} {
		ctx := context.Background()

		// Test cases:
		//  0) Test that CREATE ROLE ... NOLOGIN ... doesn't show up.
		// 	1) Test that CREATE ROLE ... LOGIN ... results in a user that GetUserListing() returns
		// 	2) Test that CREATE ROLE ... LOGIN ... IN ROLE ... results in a user that GetUserListing() returns and is part of the
		// 	   specified group (role).
		// 	3) Test that CREATE ROLE ... LOGIN ... SUPERUSER/CREATEDB/CREATEROLE/REPLICATION results in the proper flags being set.
		// 	4) Test granting SELECT/INSERT/UPDATE/DELETE/TRUNCATE/REFERENCES/TRIGGER privileges to tables.
		// 	5) Test granting EXECUTE privileges to routines.
		// 	6) Test granting USAGE privileges to COLLATION/DOMAIN/FOREIGN DATA WRAPPER/FOREIGN SERVER/SEQUENCE.
		// 	7) Test being able to identify and flatten multi-level inheritance (e.g. A -> B -> C should in our case should that C
		// 	   inherits from C and C also inherits from A).
		container, db := psql_utility.SetupPostgreSQLDatabase(t, version, func(db *sqlx.DB) error {
			tx := db.MustBegin()

			// Test Case #0
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_0 NOLOGIN
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #1
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_1 LOGIN PASSWORD NULL
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #2
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_2 LOGIN PASSWORD NULL
						IN ROLE test_user_1
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #3
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_3 LOGIN PASSWORD NULL
						SUPERUSER
						CREATEDB
						CREATEROLE
						REPLICATION
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #4 - Need to test giving the permission directly to the user AND to a parent role.
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_4a LOGIN PASSWORD NULL
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					CREATE ROLE test_user_4b LOGIN PASSWORD NULL
						IN ROLE test_user_4a
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					CREATE TABLE test_table (
						id BIGSERIAL PRIMARY KEY
					)
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					GRANT SELECT, INSERT, UPDATE ON test_table TO test_user_4a
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					GRANT DELETE, TRUNCATE, REFERENCES, TRIGGER ON test_table TO test_user_4b
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #5
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_5 LOGIN PASSWORD NULL
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					CREATE FUNCTION test5() RETURNS void AS $$
						BEGIN
							RAISE NOTICE 'Hello!';
						END;
					$$ LANGUAGE plpgsql
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					GRANT EXECUTE ON FUNCTION test5() TO test_user_5
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #6
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_6 LOGIN PASSWORD NULL
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					CREATE SEQUENCE test6 START 101;
				`); err != nil {
					tx.Rollback()
					return err
				}

				if _, err := tx.Exec(`
					GRANT USAGE ON SEQUENCE test6 TO test_user_6
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			// Test Case #7
			{
				if _, err := tx.Exec(`
					CREATE ROLE test_user_7 LOGIN PASSWORD NULL
						IN ROLE test_user_2
				`); err != nil {
					tx.Rollback()
					return err
				}
			}

			return tx.Commit()
		})
		defer container.Terminate(ctx)

		expectedUsers := map[string]*types.EtlUser{
			"test_user_1": &types.EtlUser{
				Username: "test_user_1",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self": []string{},
						},
					},
				},
			},
			"test_user_2": &types.EtlUser{
				Username: "test_user_2",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self": []string{},
						},
					},
					"test_user_1": &types.EtlRole{
						Name:        "test_user_1",
						Permissions: map[string][]string{},
					},
				},
			},
			"test_user_3": &types.EtlUser{
				Username: "test_user_3",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self": []string{PsqlSuperPermission, PsqlCreateRolePermission, PsqlCreateDbPermission, PsqlReplicationPermission},
						},
					},
				},
			},
			"test_user_4a": &types.EtlUser{
				Username: "test_user_4a",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self":                                  []string{},
							"TBL::postgres.public.test_table":       []string{"SELECT", "INSERT", "UPDATE"},
							"COLUMN::postgres.public.test_table.id": []string{"SELECT", "INSERT", "UPDATE"},
						},
					},
				},
			},
			"test_user_4b": &types.EtlUser{
				Username: "test_user_4b",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self":                                  []string{},
							"TBL::postgres.public.test_table":       []string{"DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"},
							"COLUMN::postgres.public.test_table.id": []string{"REFERENCES"},
						},
					},
					"test_user_4a": &types.EtlRole{
						Name: "test_user_4a",
						Permissions: map[string][]string{
							"TBL::postgres.public.test_table":       []string{"SELECT", "INSERT", "UPDATE"},
							"COLUMN::postgres.public.test_table.id": []string{"SELECT", "INSERT", "UPDATE"},
						},
					},
				},
			},
			"test_user_5": &types.EtlUser{
				Username: "test_user_5",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self": []string{},
							// Hopefully this doesn't change from run to run?
							"ROUTINE::postgres.public.test5_16399": []string{"EXECUTE"},
						},
					},
				},
			},
			"test_user_6": &types.EtlUser{
				Username: "test_user_6",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self":                            []string{},
							"SEQUENCE::postgres.public.test6": []string{"USAGE"},
						},
					},
				},
			},
			"test_user_7": &types.EtlUser{
				Username: "test_user_7",
				Roles: map[string]*types.EtlRole{
					"Self": &types.EtlRole{
						Name: "Self",
						Permissions: map[string][]string{
							"Self": []string{},
						},
					},
					"test_user_1": &types.EtlRole{
						Name:        "test_user_1",
						Permissions: map[string][]string{},
					},
					"test_user_2": &types.EtlRole{
						Name:        "test_user_2",
						Permissions: map[string][]string{},
					},
				},
			},
		}

		connector, err := createPsqlConnectorUser(&databases.DB{SqlxLike: db})
		if err != nil {
			t.Error(err)
		}

		users, _, err := connector.GetUserListing()
		if err != nil {
			t.Error(err)
		}
		// The +1 is for the admin user that we'll ignore.
		g.Expect(len(users)).To(gomega.Equal(len(expectedUsers) + 1))

		for _, u := range users {
			if u.Username == "postgres" {
				continue
			}
			refU, ok := expectedUsers[u.Username]
			g.Expect(ok).To(gomega.BeTrue(), "Failed to find ref user: "+u.Username)

			g.Expect(u.Username).To(gomega.Equal(refU.Username))
			g.Expect(u.FullName).To(gomega.Equal(""))
			g.Expect(u.Email).To(gomega.Equal(""))
			g.Expect(u.CreatedTime).To(gomega.BeNil())
			g.Expect(u.LastChangeTime).To(gomega.BeNil())

			for rolName, _ := range refU.Roles {
				_, ok := u.Roles[rolName]
				g.Expect(ok).To(gomega.BeTrue(), "Failed to find parent role permissions: "+rolName)
			}

			for rolName, role := range u.Roles {
				refRole, ok := refU.Roles[rolName]
				g.Expect(ok).To(gomega.BeTrue(), "Failed to find ref role: "+rolName)

				g.Expect(role.Name).To(gomega.Equal(refRole.Name))

				// Need to do a first pass through the ref role permission keys to ensure
				// completeness of the object.
				for object, _ := range refRole.Permissions {
					_, ok := role.Permissions[object]
					g.Expect(ok).To(gomega.BeTrue(), "Failed to find permissions: "+object)
				}

				for object, permissions := range role.Permissions {
					// Ignore permissions that target the information_schema or pg_catalog
					if strings.Contains(object, "information_schema") || strings.Contains(object, "pg_catalog") {
						continue
					}

					refPermissions, ok := refRole.Permissions[object]
					g.Expect(ok).To(gomega.BeTrue(), "Failed to find ref permissions: "+object)

					sort.Strings(permissions)
					sort.Strings(refPermissions)
					g.Expect(len(permissions)).To(gomega.Equal(len(refPermissions)))
					for i, p := range permissions {
						g.Expect(p).To(gomega.Equal(refPermissions[i]))
					}
				}
			}
		}
	}
}
