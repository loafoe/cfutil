package cfutil

import (
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"net/url"
	"os"
)

func Services() (map[string]*consul.AgentService, error) {
	client, err := NewConsulClient()
	if err != nil {
		return nil, err
	}
	return client.Agent().Services()
}

func ServiceRegister(name string, path string, tags ...string) error {
	appEnv, _ := Current()
	client, err := NewConsulClient()
	if err != nil {
		return err
	}
	err = client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Name:    name,
		Address: appEnv.ApplicationURIs[0],
		Tags:    tags,
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf(schemaForServices() + "://" + appEnv.ApplicationURIs[0] + path),
			Interval: "60s",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewConsulClient() (*consul.Client, error) {
	dialScheme, dialHost, err := consulDialstring("consul")
	if err != nil {
		return nil, err
	}
	client, consulErr := consul.NewClient(&consul.Config{
		Address: dialHost,
		Scheme:  dialScheme,
		Token:   os.Getenv("CONSUL_TOKEN"),
	})
	if consulErr != nil {
		return nil, consulErr
	}
	return client, nil
}

func consulDialstring(serviceName string) (string, string, error) {
	consulMaster := ""
	if consulMaster = os.Getenv("CONSUL_MASTER"); consulMaster != "" {
		parsed, err := url.Parse(consulMaster)
		if err == nil {
			return parsed.Scheme, parsed.Host, nil
		}
	}
	return "", "", errors.New("CONSUL_MASTER not found or invalid url")
}

func schemaForServices() string {
	if ForceHTTP() {
		return "http"
	}
	return "https"
}
