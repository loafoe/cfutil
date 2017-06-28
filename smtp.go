package cfutil

import (
	"errors"
	"net/url"

	cfenv "github.com/cloudfoundry-community/go-cfenv"
)

type SMTPService struct {
	url.URL
	Authentication     string
	EnableStartTLSAuto string
}

func FindSMTPService(serviceName string) (*SMTPService, error) {
	appEnv, _ := Current()
	service := &cfenv.Service{}
	err := errors.New("")
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingServiceURN(appEnv, "smtp:")
	}
	if err != nil {
		return nil, err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return nil, errors.New("SMTP credentials could not be read")
	}
	var s SMTPService
	uri, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	s.URL = *uri
	if str, ok := service.Credentials["authentication"].(string); ok {
		s.Authentication = str
	}
	if str, ok := service.Credentials["enable_starttls_auto"].(string); ok {
		s.EnableStartTLSAuto = str
	}
	return &s, nil
}
