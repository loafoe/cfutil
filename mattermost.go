package cfutil

import (
	"errors"
	"net/url"
	"strings"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
)

func MattermostDSN(serviceName string) (*url.URL, error) {
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingServiceURN(appEnv, "mattermost:")
	}
	if err != nil {
		return nil, err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return nil, errors.New("Mattermost credentials could not be read")
	}
	return url.Parse(strings.TrimPrefix(str, "mattermost:"))
}
