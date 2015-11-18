package cfutil

import (
	"errors"
	"os"
	"strings"
)

func IsFirstInstance() bool {
	appEnv, err := Current()

	if err == nil && appEnv.Index == 0 {
		return true
	}
	return false
}

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

func ListenString() string {
	return ":" + os.Getenv("PORT")
}

func Getenv(name string) string {
	return os.Getenv(name)
}
