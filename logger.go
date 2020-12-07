package gormer

import "log"

// Logger log message
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Error(...interface{})
}

// DefaultLogger default logger, console output
type DefaultLogger struct{}

var _ Logger = &DefaultLogger{}

// Debug log a debug message
func (*DefaultLogger) Debug(v ...interface{}) {
	log.Println(v)
}

// Info log a message
func (*DefaultLogger) Info(v ...interface{}) {
	log.Println(v)
}

// Error log a error message
func (*DefaultLogger) Error(v ...interface{}) {
	log.Println(v)
}

// NoLogger no log to output
type NoLogger struct{}

var _ Logger = &NoLogger{}

// Debug log a debug message
func (*NoLogger) Debug(...interface{}) {}

// Info log a message
func (*NoLogger) Info(...interface{}) {}

// Error log a error message
func (*NoLogger) Error(...interface{}) {}
