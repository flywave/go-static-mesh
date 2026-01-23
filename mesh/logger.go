package mesh

import "log"

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type DefaultLogger struct{}

func (d *DefaultLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}

func (d *DefaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (d *DefaultLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (d *DefaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

type NoOpLogger struct{}

func (d *NoOpLogger) Debug(msg string, args ...interface{}) {}

func (d *NoOpLogger) Info(msg string, args ...interface{}) {}

func (d *NoOpLogger) Warn(msg string, args ...interface{}) {}

func (d *NoOpLogger) Error(msg string, args ...interface{}) {}
