package cfutil

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
)

func NewLogger() *logrus.Logger {
	logger := logrus.New()
	appName, _ := GetApplicationName()
	formatter := &HSDPFormatter{}
	formatter.Init(appName, "", "", "")
	logger.Formatter = formatter
	return logger
}

type HSDPFormatter struct {
	template logMessage
}

type Value struct {
	Message string `json:"string"`
}

type logMessage struct {
	App         string        `json:"app"`
	Value       Value         `json:"val"`
	Version     string        `json:"ver"`
	Event       string        `json:"evt"`
	Severity    string        `json:"sev`
	Transaction string        `json:"trns"`
	User        string        `json:"usr"`
	Server      string        `json:"srv"`
	Service     string        `json:"service"`
	Instance    string        `json:"inst"`
	Category    string        `json:"cat"`
	Component   string        `json:"cmp"`
	Time        string        `json:"time"`
	Fields      logrus.Fields `json:"fields,omitempty"`
}

func (f *HSDPFormatter) Init(app, version, instance, component string) {
	f.template.App = app
	f.template.Version = version
	f.template.Instance = instance
	f.template.Component = component
}

func (f *HSDPFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := f.template
	data.Time = entry.Time.Format(logrus.DefaultTimestampFormat)
	data.Value.Message = entry.Message
	data.Severity = entry.Level.String()

	data.Fields = make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		switch k {
		case "transaction":
			data.Transaction = v.(string)
			continue
		case "user":
			data.User = v.(string)
			continue
		}
		switch v := v.(type) {
		case error:
			data.Fields[k] = v.Error()
		default:
			data.Fields[k] = v
		}
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
