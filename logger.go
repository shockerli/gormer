package gormer

import "log"

// Logger log message
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
}

// DefaultLogger default logger, console output
type DefaultLogger struct{}

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

// Debug log a debug message
func (*NoLogger) Debug(v ...interface{}) {}

// Info log a message
func (*NoLogger) Info(v ...interface{}) {}

// Error log a error message
func (*NoLogger) Error(v ...interface{}) {}
