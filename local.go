package cfutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jeffail/gabs"
)

// Struct that simulates the Cloudfoundry application environment
type vcapApplication struct {
	ApplicationName    string    `json:"application_name"`
	ApplicationVersion string    `json:"application_version"`
	ApplicationUris    []string  `json:"application_uris"`
	Host               string    `json:"host"`
	Name               string    `json:"name"`
	InstanceID         string    `json:"instance_id"`
	InstanceIndex      int       `json:"instance_index"`
	Port               int       `json:"port"`
	Start              time.Time `json:"start"`
	StartedAt          time.Time `json:"started_at"`
	StartedTimestamp   int64     `json:"started_timestamp"`
	Uris               []string  `json:"uris"`
	Users              *[]string `json:"users"`
	Version            string    `json:"version"`
}

func localVcapApplication() string {
	appID := uuid.New().String()
	port := 8080
	host := "localhost"
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = p
	}
	hostWithPort := fmt.Sprintf("%s:%d", host, port)

	va := &vcapApplication{
		ApplicationName:    "appname",
		ApplicationVersion: appID,
		Host:               "0.0.0.0",
		Port:               port,
		ApplicationUris:    []string{hostWithPort},
		InstanceID:         "451f045fd16427bb99c895a2649b7b2a",
		InstanceIndex:      0,
		Name:               "appname",
		Start:              time.Now(),
		StartedAt:          time.Now(),
		StartedTimestamp:   time.Now().Unix(),
		Uris:               []string{hostWithPort},
		Version:            appID,
	}
	json, _ := json.Marshal(va)
	return string(json)
}

func localMemoryLimit() string {
	return "2G"
}

func localVcapServices() string {
	var supportedServices = []string{
		"postgres",
		"smtp",
		"rabbitmq",
		"sentry",
	}
	jsonObj := gabs.New()
	jsonObj.Array("user-provided")
	for _, service := range supportedServices {
		env := "CF_LOCAL_" + strings.ToUpper(service)
		uris := os.Getenv(env)
		items := strings.Split(uris, "|")
		for _, item := range items {
			if item == "" {
				continue
			}
			serviceJSON := gabs.New()
			name := service
			uri := item
			if components := strings.Split(item, ","); len(components) > 1 {
				name = components[0]
				uri = components[1]
			}
			serviceJSON.Set(name, "name")
			serviceJSON.Set(uri, "credentials", "uri")
			jsonObj.ArrayAppendP(serviceJSON.Data(), "user-provided")
		}
	}
	return jsonObj.String()
}

func IsLocal() bool {
	return os.Getenv("CF_LOCAL") == "true"
}

// ListenString() returns the listen string based on the `PORT` environment variable value
func ListenString() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
