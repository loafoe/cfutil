package cfutil

import (
	"errors"
	"net/url"
	"strings"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
)

func MattermostDSN(serviceName string) (*url.URL, error) {
	appEnv, _ := Current()
	var service *cfenv.Service
	var err error
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingServiceURN(appEnv, "mattermost:")
	}
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, errors.New("No mattermost service found")
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return nil, errors.New("Mattermost credentials could not be read")
	}
	return url.Parse(strings.TrimPrefix(str, "mattermost:"))
}
