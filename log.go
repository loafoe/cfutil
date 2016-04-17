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
	Message string `json:"message"`
}

type logMessage struct {
	App         string        `json:"app"`
	Value       Value         `json:"val"`
	Version     string        `json:"ver,omitempty"`
	Event       string        `json:"evt,omitempty"`
	Severity    string        `json:"sev,omitempty"`
	Transaction string        `json:"trns,omitempty"`
	User        string        `json:"usr,omitempty"`
	Server      string        `json:"sr,omitemptyv"`
	Service     string        `json:"service,omitempty"`
	Instance    string        `json:"inst,omitempty"`
	Category    string        `json:"cat,omitempty"`
	Component   string        `json:"cmp,omitempty"`
	Time        string        `json:"time,omitempty"`
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
