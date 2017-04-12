package cfutil

import (
	"errors"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"strings"
)

type PHService struct {
	Type            string
	ApplicationName string
	PropositionName string
	BaseURL         string
	Key             string
	Secret          string
}

func ConnectPHService(serviceName string) (*PHService, error) {
	var phService PHService
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	service, err = firstMatchingServiceURN(appEnv, serviceName+":")
	if err != nil {
		return nil, err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return nil, errors.New("Connect PH service  could not be read")
	}
	splitted := strings.Split(":", str)
	if len(splitted) != 6 {
		return nil, errors.New("Expected ph[div]:appName:propName:key:secret:baseURI")
	}
	if splitted[0] != serviceName {
		return nil, errors.New("URN mismatch: " + splitted[0])
	}
	phService.ApplicationName = splitted[1]
	phService.PropositionName = splitted[2]
	phService.Key = splitted[3]
	phService.Secret = splitted[4]
	phService.BaseURL = splitted[5]
	return &phService, nil
}
