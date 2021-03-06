// Package cfutil provides a number of convenience functions to apps
// being developed for use in a Cloudfoundry environment. For local
// development it can simulate the CF environment variables that your
// app typically expects. It can also emulate Cloudfoundry services
// based on environment variables.
package cfutil

import (
	"errors"
	"fmt"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

//
// NewConnection opens a new database connection given the driver and CF service name.
// Currently only postgres is supported.
//
func NewConnection(driver, name string) (conn *sqlx.DB, connectString string, err error) {
	appEnv, _ := Current()
	switch driver {
	case "postgres":
		if name == "" {
			connectString, err = firstMatchingServiceURI(appEnv, driver)
		} else {
			connectString, err = postgresConnectString(appEnv, name)
		}
	default:
		return nil, "", fmt.Errorf("Unsupported driver '%s'", driver)
	}
	if err != nil {
		return nil, "", err
	}

	conn, err = sqlx.Connect(driver, connectString)
	return
}

func postgresConnectString(env *cfenv.App, serviceName string) (string, error) {
	postgresService, err := env.Services.WithName(serviceName)
	if err != nil {
		return "", err
	}
	str, ok := postgresService.Credentials["uri"].(string)
	if !ok {
		return "", errors.New("Postgres credentials not available")
	}
	return str, nil
}
