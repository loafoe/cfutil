package cfutil

import (
	"errors"
	"net/url"
	"strings"

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
	str = strings.TrimPrefix(str, `smtp://`)
	s.Scheme = `smtp`
	splitted := strings.Split(str, `@`)
	if len(splitted) > 1 {
		s.Host = splitted[1]
	}
	userPass := strings.Split(splitted[0], `:`)
	if len(userPass) > 1 {
		s.User = url.UserPassword(userPass[0], userPass[1])
	} else {
		s.User = url.UserPassword(userPass[0], "")
	}
	if str, ok := service.Credentials["authentication"].(string); ok {
		s.Authentication = str
	}
	if str, ok := service.Credentials["enable_starttls_auto"].(string); ok {
		s.EnableStartTLSAuto = str
	}
	return &s, nil
}
