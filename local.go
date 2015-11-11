package cfutil

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jeffail/gabs"
	"github.com/satori/go.uuid"
	"os"
	"strconv"
	"strings"
	"time"
)

type vcapApplication struct {
	ApplicationName    string    `json:"application_name"`
	ApplicationVersion string    `json:"application_version"`
	ApplicationUris    []string  `json:"application_uris"`
	Host               string    `json:"host"`
	Name               string    `json:"name"`
	InstanceId         string    `json:"instance_id"`
	InstanceIndex      int       `json:"instance_index"`
	Port               int       `json:"port"`
	Start              time.Time `json:"start"`
	StartedAt          time.Time `json:"started_at"`
	StartedTimestamp   int64     `json:"started_timestamp"`
	Uris               []string  `json:"uris"`
	Users              *[]string `json:"users"`
	Version            string    `json:"version"`
}

func LocalVcapApplication() string {
	appId := uuid.NewV4().String()
	port := 8080
	host := "localhost"
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = p
	}
	hostWithPort := fmt.Sprintf("%s:%d", host, port)

	va := &vcapApplication{
		ApplicationName:    "appname",
		ApplicationVersion: appId,
		Host:               "0.0.0.0",
		Port:               port,
		ApplicationUris:    []string{hostWithPort},
		InstanceId:         "451f045fd16427bb99c895a2649b7b2a",
		InstanceIndex:      0,
		Name:               "appname",
		Start:              time.Now(),
		StartedAt:          time.Now(),
		StartedTimestamp:   time.Now().Unix(),
		Uris:               []string{hostWithPort},
		Version:            appId,
	}
	json, _ := json.Marshal(va)
	return string(json)
}

func LocalMemoryLimit() string {
	return "2G"
}

func LocalVcapServices() string {
	var supportedServices = []string{
		"postgres",
		"smtp",
		"rabbitmq",
		"consul",
	}
	jsonObj := gabs.New()
	for _, service := range supportedServices {
		env := "CF_LOCAL_" + strings.ToUpper(service)
		uri := os.Getenv(env)
		if uri != "" {
			jsonObj.Array(service)
			serviceJson := gabs.New()
			name := service
			if components := strings.Split(uri, ","); len(components) > 1 {
				name = components[0]
				uri = components[1]
			}
			serviceJson.Set(name, "name")
			serviceJson.Set(uri, "credentials", "uri")
			log.Print("Local service: ", name)
			jsonObj.ArrayAppendP(serviceJson.Data(), service)
		}
	}
	return jsonObj.String()
}
