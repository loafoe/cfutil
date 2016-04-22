package cfutil

import (
	"errors"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
)

func SentryDSN(serviceName string) (string, error) {
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingServiceByCredential(appEnv, "sentry_dsn")
	}
	if err != nil {
		return "", err
	}
	str, ok := service.Credentials["sentry_dsn"].(string)
	if !ok {
		return "", errors.New("Sentry credentials could not be read")
	}
	return str, nil
}
