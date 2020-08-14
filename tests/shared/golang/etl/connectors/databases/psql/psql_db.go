package psql_utility

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

type InitFunction func(*sqlx.DB) error

func SetupPostgreSQLDatabase(t *testing.T, version string, fns ...InitFunction) (testcontainers.Container, *sqlx.DB) {
	ctx := context.Background()
	waitStrat := wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(1 * time.Minute)
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("postgres:%s", version),
		ExposedPorts: []string{"5432"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
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
		db, err = sqlx.Connect("postgres", fmt.Sprintf("postgres://postgres:password@%s?sslmode=disable", endpoint))
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
