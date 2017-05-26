package cfutil

import (
	"errors"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"strings"
)

type PHService struct {
	Type            string `json:"type"`
	ApplicationName string `json:"application_name"`
	PropositionName string `json:"proposition_name"`
	BaseURL         string `json:"base_url"`
	SharedKey       string `json:"shared_key"`
	SharedSecret    string `json:"shared_secret"`
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
	phService.BaseURL = strings.TrimPrefix(str, phServiceType+":")
	if val, ok := service.Credentials["application_name"].(string); ok {
		phService.ApplicationName = val
	}
	if val, ok := service.Credentials["proposition_name"].(string); ok {
		phService.PropositionName = val
	}
	if val, ok := service.Credentials["shared_key"].(string); ok {
		phService.SharedKey = val
	}
	if val, ok := service.Credentials["shared_secret"].(string); ok {
		phService.SharedSecret = val
	}
	phService.Type = phServiceType
	return &phService, nil
}
