// Logger code adapted from: https://github.com/euroteltr/rbac
// Copyright (c) 2018 Eurotel AS (www.eurotel.com.tr)
// Licensed under the MIT License

package rbac

import "fmt"

// Logger defines logging behavior for RBAC library
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NullLogger implements Logger and performs no logging
type NullLogger struct{}

// NewNullLogger returns a no-op logger (default if none provided)
func NewNullLogger() Logger {
	return &NullLogger{}
}

func (l *NullLogger) Debugf(format string, args ...interface{}) {}
func (l *NullLogger) Errorf(format string, args ...interface{}) {}

// ConsoleLogger logs to stdout (useful for development)
type ConsoleLogger struct{}

// NewConsoleLogger returns a simple stdout logger
func NewConsoleLogger() Logger {
	return &ConsoleLogger{}
}

func (l *ConsoleLogger) Debugf(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}

func (l *ConsoleLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
