package cfutil

import (
	"errors"
	"os"
	"strings"
)

// IsFirstInstance() returns true if the instance calling it
// is the first running instance of the app within Cloudfoundry.
// This is useful if you want to for example trigger database
// migrations but only want to execute these on the first starting instance.
func IsFirstInstance() bool {
	appEnv, err := Current()

	if err == nil && appEnv.Index == 0 {
		return true
	}
	return false
}

// GetApplicationName() returns the name of the app within Cloudfoundry
func GetApplicationName() (string, error) {
	appEnv, err := Current()
	if err != nil {
		return "unknown", err
	}
	return appEnv.Name, nil
}

// GetHostname() returns the (first) hostname designated to your app within Cloudfoundry
func GetHostname() (string, error) {
	appEnv, err := Current()
	if err != nil {
		return "", err
	}
	if len(appEnv.ApplicationURIs) == 0 {
		return "", errors.New("Unknown hostname")
	}
	hostWithPort := appEnv.ApplicationURIs[0]
	parts := strings.Split(hostWithPort, ":")
	return parts[0], nil
}

// Getenv(name) returns the environment variable value of ENV `name`
func Getenv(name string) string {
	return os.Getenv(name)
}
