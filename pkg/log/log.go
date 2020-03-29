package log

import "log"

// Logger is logging interface of gomelift.
type Logger interface {
	Log(msg string, args ...interface{})
	Panic(msg string, args ...interface{})
}

type StandardLogger struct {
}

func (logger *StandardLogger) Log(msg string, args ...interface{}) {
	log.Println(msg, args)
}

func (logger *StandardLogger) Panic(msg string, args ...interface{}) {
	log.Panic(msg, args)
}
