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

func ConnectPHService(phServiceType, serviceName string) (*PHService, error) {
	var phService PHService
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	service, err = appEnv.Services.WithName(serviceName)
	if err != nil {
		service, err = firstMatchingServiceURN(appEnv, phServiceType+":")
	}
	if err != nil {
		return nil, err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return nil, errors.New("Connect PH service  could not be read")
	}
	splitted := strings.Split(":", str)
	if len(splitted) != 7 { // because last field will be split on ://
		return nil, errors.New("Expected ph[div]:appName:propName:key:secret:baseURI")
	}
	if splitted[0] != phServiceType {
		return nil, errors.New("URN mismatch: " + splitted[0])
	}
	phService.ApplicationName = splitted[1]
	phService.PropositionName = splitted[2]
	phService.Key = splitted[3]
	phService.Secret = splitted[4]
	phService.BaseURL = strings.Join(splitted[5:], ":")
	return &phService, nil
}
