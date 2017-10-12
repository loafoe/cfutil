package cfutil

import (
	"errors"
	"fmt"
	"strings"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
)

type PHService struct {
	Type            string `json:"type"`
	ApplicationName string `json:"application_name"`
	PropositionName string `json:"proposition_name"`
	BaseURL         string `json:"base_url"`
	SharedKey       string `json:"shared_key"`
	SharedSecret    string `json:"shared_secret"`
	Client          string `json:"client"`
	Password        string `json:"password"`
}

func ConnectPHService(phServiceType, serviceName string) (*PHService, error) {
	var phService PHService
	appEnv, _ := Current()
	var service *cfenv.Service
	var err error
	service, err = appEnv.Services.WithName(serviceName)
	if err != nil {
		service, err = firstMatchingServiceURN(appEnv, phServiceType+":")
		if err != nil {
			return nil, err
		}
	}
	if service == nil {
		return nil, fmt.Errorf("Connect PH service not found: %s", phServiceType)
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return nil, errors.New("Connect PH service could not be read")
	}
	if !strings.HasPrefix(str, phServiceType+":") {
		return nil, fmt.Errorf("PH service mismatch: %s --> %s", phServiceType, str)
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
