package cfutil

import (
	"errors"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"strings"
)

func SentryDSN(serviceName string) (string, error) {
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingServiceURN(appEnv, "sentry:")
	}
	if err != nil {
		return "", err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return "", errors.New("Sentry credentials could not be read")
	}
	return strings.TrimPrefix(str, "sentry:"), nil
}
