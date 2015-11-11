package cfutil

import (
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"net/url"
)

func ServiceRegister(name string, tags ...string) error {
	appEnv, _ := Current()
	client, err := NewConsulClient()
	if err != nil {
		return err
	}
	err = client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Name:    name,
		Address: appEnv.ApplicationURIs[0],
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf(schemaForConsul() + "://" + appEnv.ApplicationURIs[0] + "/health"),
			Interval: "30s",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewConsulClient() (*consul.Client, error) {
	dialString, err := consulDialstring("consul")
	if err != nil {
		return nil, err
	}
	client, consulErr := consul.NewClient(&consul.Config{
		Address: dialString,
		Scheme:  schemaForConsul(),
	})
	if consulErr != nil {
		return nil, consulErr
	}
	return client, nil
}

func consulDialstring(serviceName string) (string, error) {
	appEnv, _ := Current()
	consulService, err := appEnv.Services.WithName(serviceName)
	if err != nil {
		return "", err
	}
	port := "8500"

	uri, ok := consulService.Credentials["uri"].(string)
	if ok {
		u, err := url.Parse(uri)
		if err == nil {
			return fmt.Sprintf("%s", u.Host), nil
		}
		// Fallback to hostname/port lookup
	}

	hostname, ok := consulService.Credentials["hostname"].(string)
	if !ok {
		return "", errors.New("consul service not available")
	}
	port, ok = consulService.Credentials["port"].(string)
	if ok {
		return fmt.Sprintf("%s:%s", hostname, port), nil
	}
	return fmt.Sprintf("%s:8500", hostname), nil
}

func schemaForConsul() string {
	return "http"
}
