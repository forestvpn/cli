package auth

import (
	"fmt"
	"log"
	"strings"

	"github.com/forestvpn/cli/utils"
	"github.com/forestvpn/goauthlib/pkg/logger"
	"github.com/forestvpn/goauthlib/pkg/svc"
	"github.com/sirupsen/logrus"
)

var AuthStore, _ = svc.NewFilePersistentStore()

// SimpleLogger implements the Logger interface using the Go standard library's log package
type SimpleLogger struct {
	fields logrus.Fields
}

func (l *SimpleLogger) Debugf(format string, args ...interface{}) {
	if utils.Verbose {
		utils.InfoLogger.Println(l.renderLogString(format, args...))
	}
}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	utils.InfoLogger.Println(l.renderLogString(format, args...))
}

func (l *SimpleLogger) Printf(format string, args ...interface{}) {
	utils.InfoLogger.Println(l.renderLogString(format, args...))
}

func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	utils.InfoLogger.Println(l.renderLogString(format, args...))
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	utils.InfoLogger.Println(l.renderLogString(format, args...))
}

func (l *SimpleLogger) Fatalf(format string, args ...interface{}) {
	log.Fatalln(l.renderLogString(format, args...))
}

func (l *SimpleLogger) Panicf(format string, args ...interface{}) {
	log.Fatalln(l.renderLogString(format, args...))
}

func (l *SimpleLogger) WithField(key string, value interface{}) logger.Logger {
	newFields := l.cloneFields(l.fields)
	newFields[key] = value
	return &SimpleLogger{
		fields: newFields,
	}
}

func (l *SimpleLogger) WithFields(fields logrus.Fields) logger.Logger {
	newFields := l.cloneFields(l.fields)
	for k, v := range fields {
		newFields[k] = v
	}
	return &SimpleLogger{
		fields: newFields,
	}
}

func (l *SimpleLogger) WithError(err error) logger.Logger {
	newFields := l.cloneFields(l.fields)
	newFields["error"] = err
	return &SimpleLogger{
		fields: newFields,
	}
}

func (l *SimpleLogger) cloneFields(fields logrus.Fields) logrus.Fields {
	newFields := make(logrus.Fields, len(fields))
	for k, v := range fields {
		newFields[k] = v
	}
	return newFields
}
func (l *SimpleLogger) renderFields() string {
	preRendered := make([]string, 0, len(l.fields))
	for k, v := range l.fields {
		preRendered = append(preRendered, fmt.Sprintf("%s=%+v", k, v))
	}
	return fmt.Sprintf("[%s]", strings.Join(preRendered, ", "))
}

func (l *SimpleLogger) renderLogString(format string, args ...interface{}) string {
	prefix := fmt.Sprintf("%s", l.renderFields())
	return prefix + ": " + fmt.Sprintf(format, args...)
}

func NewSimpleLogger() logger.Logger {
	return &SimpleLogger{}
}

func AuthService(userID string) svc.Svc {
	return svc.New(userID,
		svc.WithAuthSvcBaseUrl(utils.ApiHost[4:]),
		svc.WithAuthSvcLogger(NewSimpleLogger()),
		svc.WithAuthSvcAutoOpen(true),
		svc.WithAuthSvcPersistentStore(AuthStore),
	)
}
