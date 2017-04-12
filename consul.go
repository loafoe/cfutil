package cfutil

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

// Services() returns the list of services available from the
// Consul cluster
func Services() ([]string, error) {
	client, err := NewConsulClient()
	if err != nil {
		return nil, err
	}
	catalogServices, _, err := client.Catalog().Services(nil)
	if err != nil {
		return []string{}, err
	}
	var services []string
	for k := range catalogServices {
		services = append(services, k)
	}
	return services, nil
}

func DiscoverServiceURL(serviceName, tags string) (string, error) {
	client, err := NewConsulClient()
	if err != nil {
		return "", err
	}
	services, _, err := client.Catalog().Service(serviceName, tags, nil)
	if err != nil {
		return "", fmt.Errorf("Service `%s` not found: %s", serviceName, err)
	}
	if len(services) > 0 {
		return CreateURLFromServiceCatalog(services[0])
	}
	return "", fmt.Errorf("Service `%s` not found", serviceName)

}

func CreateURLFromServiceCatalog(catalog *consul.CatalogService) (string, error) {
	var serviceURL url.URL
	if catalog.ServicePort == 443 {
		serviceURL.Scheme = "https"
		serviceURL.Host = catalog.ServiceAddress
	} else {
		serviceURL.Scheme = "http"
		serviceURL.Host = fmt.Sprintf("%s:%d", catalog.ServiceAddress, catalog.ServicePort)
	}
	return serviceURL.String(), nil
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

	appURL, _ := url.Parse(schema + "://" + appEnv.ApplicationURIs[0])
	splitted := strings.Split(appURL.Host, ":")
	hostWithoutPort := splitted[0]
	if hostWithoutPort == "" {
		hostWithoutPort = "localhost"
	}
	if len(splitted) > 1 {
		addedPort, err := strconv.Atoi(splitted[1])
		if err == nil && addedPort != port {
			port = addedPort
		}
	}

	err = client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		Name:    name,
		Address: hostWithoutPort,
		Port:    port,
		Tags:    tags,
		Check: &consul.AgentServiceCheck{
			HTTP:     fmt.Sprintf(schema + "://" + appURL.Host + path),
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

func GetConsulKey(mooncoreKey string) (string, error) {
	ns := ConsulNamespace()
	key := "mooncore/" + ns + "/" + mooncoreKey
	client, err := NewConsulClient()
	if err != nil {
		return "", err
	}
	kvPair, _, err := client.KV().Get(key, nil)
	if err != nil {
		return "", err
	}
	if kvPair == nil || kvPair.Value == nil {
		return "", fmt.Errorf("Key not found: %s", mooncoreKey)
	}
	return string(kvPair.Value), nil
}

func ConsulDatacenter() (string, error) {
	client, err := NewConsulClient()
	if err != nil {
		return "", err
	}
	self, err := client.Agent().Self()
	if err != nil {
		return "", err
	}
	dc, ok := self["Config"]["Datacenter"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid lookup for Datacenter")
	}
	return dc, nil
}

func ConsulNamespace() string {
	return os.Getenv("CONSUL_NAMESPACE")
}

func consulDialstring(serviceName string) (string, string, error) {
	consulMaster := ""
	if consulMaster = os.Getenv("CONSUL_MASTER"); consulMaster != "" {
		parsed, err := url.Parse(consulMaster)
		if err == nil {
			return parsed.Scheme, parsed.Host, nil
		}
	}
	return "", "", fmt.Errorf("CONSUL_MASTER not found or invalid url: %s", consulMaster)
}

func schemaAndPortForServices() (string, int) {
	if ForceHTTP() {
		return "http", 80
	}
	return "https", 443
}
