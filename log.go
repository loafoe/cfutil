package cfutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
)

type Logger interface {
	Debug(c context.Context, format string, args ...interface{})
	Info(c context.Context, format string, args ...interface{})
	Warning(c context.Context, format string, args ...interface{})
	Error(c context.Context, format string, args ...interface{})
	Critical(c context.Context, format string, args ...interface{})
	Raw(c context.Context, rawMessage string)
}

func NewLogger() Logger {
	newLogger := HSDPLogger{}
	appName, _ := GetApplicationName()
	newLogger.Init(appName, "", "", "")
	return newLogger
}

var log = NewLogger()

type HSDPLogger struct {
	logger   *logrus.Logger
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
	Server      string        `json:"srv,omitempty"`
	Service     string        `json:"service,omitempty"`
	Instance    string        `json:"inst,omitempty"`
	Category    string        `json:"cat,omitempty"`
	Component   string        `json:"cmp,omitempty"`
	Time        string        `json:"time,omitempty"`
	Fields      logrus.Fields `json:"fields,omitempty"`
}

func (f *HSDPLogger) Init(app, version, instance, component string) {
	f.logger = logrus.New()
	f.logger.Formatter = f
	f.logger.Out = os.Stdout

	f.template.App = app
	f.template.Version = version
	f.template.Instance = instance
	if f.template.Instance == "" {
		f.template.Instance = "not-specified"
	}
	f.template.Component = component
	f.template.Category = "Tracelog"
	f.template.Event = "1"
	f.template.Server = "not-set"
	f.template.Service = "not-set"
	f.template.User = "not-specified"
}

const KeyCorrelationID = "correlationid"

func correlationIDFromContext(c context.Context) string {
	if id, ok := c.Value(KeyCorrelationID).(string); ok {
		return id
	}
	return ""
}

func (f HSDPLogger) Raw(c context.Context, rawString string) {
	fmt.Print(rawString)
}

func (f HSDPLogger) Debug(c context.Context, format string, args ...interface{}) {
	f.logger.WithField(KeyCorrelationID, correlationIDFromContext(c)).Debugf(format, args...)
}

func (f HSDPLogger) Info(c context.Context, format string, args ...interface{}) {
	f.logger.WithField(KeyCorrelationID, correlationIDFromContext(c)).Infof(format, args...)
}

func (f HSDPLogger) Warning(c context.Context, format string, args ...interface{}) {
	f.logger.WithField(KeyCorrelationID, correlationIDFromContext(c)).Warningf(format, args...)
}

func (f HSDPLogger) Error(c context.Context, format string, args ...interface{}) {
	f.logger.WithField(KeyCorrelationID, correlationIDFromContext(c)).Errorf(format, args...)
}

func (f HSDPLogger) Critical(c context.Context, format string, args ...interface{}) {
	f.logger.WithField(KeyCorrelationID, correlationIDFromContext(c)).Fatalf(format, args...)
}

func (f *HSDPLogger) Format(entry *logrus.Entry) ([]byte, error) {
	data := f.template
	data.Time = entry.Time.Format("2006-01-02T15:04:05.000Z07:00")
	data.Value.Message = entry.Message
	data.Severity = entry.Level.String()

	data.Fields = make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		switch k {
		case "transaction", KeyCorrelationID:
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
