package cfutil

import (
	"errors"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"net/url"
	"os"
)

// Services() returns the list of services available from the
// Consul cluster
func Services() (map[string]*consul.AgentService, error) {
	client, err := NewConsulClient()
	if err != nil {
		return nil, err
	}
	return client.Agent().Services()
}

// Use ServiceRegister() to register your app in the Consul cluster
// Optionally you can provide a health endpoint on your URL and
// a number of tags to make your service more discoverable
func ServiceRegister(name string, path string, tags ...string) error {
	appEnv, _ := Current()
	client, err := NewConsulClient()
	if err != nil {
		return err
	}
	schema, port := schemaAndPortForServices()
	err = client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Name:    name,
		Address: appEnv.ApplicationURIs[0],
		Port:    port,
		Tags:    tags,
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf(schema + "://" + appEnv.ApplicationURIs[0] + path),
			Interval: "60s",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// NewConsulClient() returns a new consul client which you can use to
// access the Consul cluster HTTP API. It uses `CONSUL_MASTER` and
// `CONSUL_TOKEN` environment variables to set up the HTTP API connection.
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

func schemaAndPortForServices() (string, int) {
	if ForceHTTP() {
		return "http", 80
	}
	return "https", 443
}
