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
	Username           string
	Password           string
}

func FindSMTPService(serviceName string) (*SMTPService, error) {
	appEnv, _ := Current()
	var service *cfenv.Service
	var err error
	if serviceName != "" {
		service, err = serviceByName(appEnv, serviceName)
	} else {
		service, err = firstMatchingServiceURN(appEnv, "smtp:")
	}
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, errors.New("SMTP service not found")
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
		s.Username = userPass[0]
		s.Password = userPass[1]
	} else {
		s.User = url.UserPassword(userPass[0], "")
		s.Username = userPass[0]
	}
	if str, ok := service.Credentials["authentication"].(string); ok {
		s.Authentication = str
	}
	if str, ok := service.Credentials["enable_starttls_auto"].(string); ok {
		s.EnableStartTLSAuto = str
	}
	return &s, nil
}
