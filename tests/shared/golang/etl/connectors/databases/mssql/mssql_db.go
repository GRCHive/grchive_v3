package mssql_utility

import (
	"context"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

type InitFunction func(*sqlx.DB) error

func SetupMssqlDatabase(t *testing.T, version string, fns ...InitFunction) (testcontainers.Container, *sqlx.DB) {
	ctx := context.Background()
	waitStrat := wait.ForLog("SQL Server is now ready for client connections")
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("mcr.microsoft.com/mssql/server:%s", version),
		ExposedPorts: []string{"1433"},
		Env: map[string]string{
			"SA_PASSWORD": "aaaaBBBB1111!",
			"ACCEPT_EULA": "Y",
		},
		WaitingFor: waitStrat,
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Error(err)
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}

	var db *sqlx.DB

	for {
		db, err = sqlx.Connect("sqlserver", fmt.Sprintf("sqlserver://SA:aaaaBBBB1111!@%s", endpoint))
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	for _, fn := range fns {
		err = fn(db)
		if err != nil {
			t.Error(err)
		}
	}

	return container, db
}
