package mysql_utility

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

type InitFunction func(*sqlx.DB) error

func SetupMySQLDatabase(t *testing.T, version string, fns ...InitFunction) (testcontainers.Container, *sqlx.DB) {
	ctx := context.Background()
	waitStrat := wait.ForLog("ready for connections").WithOccurrence(2).WithStartupTimeout(1 * time.Minute)
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("mysql:%s", version),
		ExposedPorts: []string{"3306"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
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
		db, err = sqlx.Connect("mysql", fmt.Sprintf("root:password@(%s)/", endpoint))
		if err == nil {
			db.SetConnMaxLifetime(time.Second)
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
