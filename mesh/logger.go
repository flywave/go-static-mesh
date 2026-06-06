package mesh

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type NoOpLogger struct{}

func (d *NoOpLogger) Debug(msg string, args ...interface{}) {}

func (d *NoOpLogger) Info(msg string, args ...interface{}) {}

func (d *NoOpLogger) Warn(msg string, args ...interface{}) {}

func (d *NoOpLogger) Error(msg string, args ...interface{}) {}
