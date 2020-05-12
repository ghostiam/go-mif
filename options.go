package mif

// https://sagikazarmark.hu/blog/functional-options-on-steroids/

type Option interface {
	apply(m *mif)
}

// Logger
type Logger interface {
	Error(args ...interface{})
}

type withLogger struct {
	logger Logger
}

func (w withLogger) apply(m *mif) {
	m.logger = w.logger
}

func WithLogger(logger Logger) Option {
	return withLogger{logger}
}

// Std Logger
type StdLogger interface {
	Println(args ...interface{})
}

type stdLoggerWrapper struct {
	logger StdLogger
}

func (s stdLoggerWrapper) Error(args ...interface{}) {
	s.logger.Println(args...)
}

func WithStdLogger(logger StdLogger) Option {
	return withLogger{stdLoggerWrapper{logger}}
}

// Noop Logger
type noopLogger struct{}

func (noopLogger) Error(_ ...interface{}) {}

func WithNoopLogger() Option {
	return withLogger{noopLogger{}}
}

// JSON Config
type JSONConfig struct {
	Prefix string
	Indent string
}

type withJSONConfig struct {
	json JSONConfig
}

func (w withJSONConfig) apply(m *mif) {
	m.json = w.json
}

func WithJSONConfig(cfg JSONConfig) Option {
	return withJSONConfig{cfg}
}

// Disable panic
type setDisablePanic struct{}

func (w setDisablePanic) apply(m *mif) {
	m.disablePanic = true
}

func SetDisablePanic() Option {
	return setDisablePanic{}
}
