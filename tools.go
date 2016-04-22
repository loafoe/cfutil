package cfutil

import (
	"errors"
	"fmt"
	cfenv "github.com/cloudfoundry-community/go-cfenv"
	"os"
	"regexp"
	"strings"
)

// IsFirstInstance() returns true if the instance calling it
// is the first running instance of the app within Cloudfoundry.
// This is useful if you want to for example trigger database
// migrations but only want to execute these on the first starting instance.
func IsFirstInstance() bool {
	appEnv, err := Current()

	if err == nil && appEnv.Index == 0 {
		return true
	}
	return false
}

// GetApplicationName() returns the name of the app within Cloudfoundry
func GetApplicationName() (string, error) {
	appEnv, err := Current()
	if err != nil {
		return "unknown", err
	}
	return appEnv.Name, nil
}

// GetHostname() returns the (first) hostname designated to your app within Cloudfoundry
func GetHostname() (string, error) {
	appEnv, err := Current()
	if err != nil {
		return "", err
	}
	if len(appEnv.ApplicationURIs) == 0 {
		return "", errors.New("Unknown hostname")
	}
	hostWithPort := appEnv.ApplicationURIs[0]
	parts := strings.Split(hostWithPort, ":")
	return parts[0], nil
}

// Getenv(name) returns the environment variable value of ENV `name`
func Getenv(name string) string {
	return os.Getenv(name)
}

func serviceByName(env *cfenv.App, serviceName string) (*cfenv.Service, error) {
	service, err := env.Services.WithName(serviceName)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func serviceURIByName(env *cfenv.App, serviceName string) (string, error) {
	service, err := serviceByName(env, serviceName)
	if err != nil {
		return "", err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return "", errors.New("Service credentials not available")
	}
	return str, nil
}

func firstMatchingService(env *cfenv.App, schema string) (*cfenv.Service, error) {
	regex, err := regexp.Compile("^" + schema + "://")
	if err != nil {
		return nil, err
	}
	for _, services := range env.Services {
		for _, service := range services {
			str, ok := service.Credentials["uri"].(string)
			if !ok {
				continue
			}
			if regex.MatchString(str) {
				return &service, nil
			}
		}
	}
	return nil, fmt.Errorf("No matching service found for '%s'", schema)
}

func firstMatchingServiceURI(env *cfenv.App, schema string) (string, error) {
	service, err := firstMatchingService(env, schema)
	if err != nil {
		return "", err
	}
	str, ok := service.Credentials["uri"].(string)
	if !ok {
		return "", errors.New("Service credentials not available")
	}
	return str, nil
}

func firstMatchingServiceByCredential(env *cfenv.App, credential string) (*cfenv.Service, error) {
	for _, services := range env.Services {
		for _, service := range services {
			if service.Credentials[credential] != nil {
				return &service, nil
			}
		}
	}
	return nil, fmt.Errorf("No matching service found that contains credential '%s'", credential)
}
