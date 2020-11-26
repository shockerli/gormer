package gormer

import "log"

// Logger log message
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
}

// default logger
type defaultLogger struct{}

func (*defaultLogger) Debug(v ...interface{}) {
	log.Println(v)
}

func (*defaultLogger) Info(v ...interface{}) {
	log.Println(v)
}

func (*defaultLogger) Error(v ...interface{}) {
	log.Println(v)
}
