package oracle_utility

import (
	"context"
	"fmt"
	_ "github.com/godror/godror"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type InitFunction func(*sqlx.DB) error

func SetupOracleDatabase(t *testing.T, version string, fns ...InitFunction) (testcontainers.Container, *sqlx.DB) {
	ctx := context.Background()
	waitStrat := wait.ForLog("DATABASE IS READY TO USE!")
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("gitlab.com/grchive/grchive-v3/oracle/database:%s", version),
		ExposedPorts: []string{"1521"},
		Env: map[string]string{
			"ORACLE_PWD":     "asdfasdf1A!",
			"ORACLE_EDITION": "enterprise",
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

	endpoint, err := container.PortEndpoint(ctx, "1521/tcp", "")
	if err != nil {
		t.Error(err)
	}

	var db *sqlx.DB

	// Copy sqlnet.ora into a temporary file with a shorter path name...the long config directory path seems to break Oracle's stuff.
	tmpDir, err := ioutil.TempDir("", "oracleCfg")
	if err != nil {
		t.Error(err)
	}

	baseDir := fmt.Sprintf("%s/grchive_v3/deps/external/oracle/instantclient_19_8", os.Getenv("TEST_SRCDIR"))
	libDir := fmt.Sprintf("%s/linux", baseDir)
	configDir := fmt.Sprintf("%s/cfg", baseDir)

	data, err := ioutil.ReadFile(configDir + "/sqlnet.ora")
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/sqlnet.ora", tmpDir), data, 0755)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 5; i++ {
		db, err = sqlx.Connect("godror", fmt.Sprintf(`
			user="system"
			password="asdfasdf1A!"
			connectString="%s/ORCLCDB"
			libDir="%s"
			configDir="%s"`, endpoint, libDir, tmpDir))
		if err == nil {
			break
		}
		fmt.Printf("ERROR %s\n", err.Error())
		time.Sleep(1 * time.Second)
		break
	}

	for _, fn := range fns {
		err = fn(db)
		if err != nil {
			t.Error(err)
		}
	}

	os.RemoveAll(tmpDir)
	return container, db
}
