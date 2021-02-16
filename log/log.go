package log

var (
	defaultLogger Logger = NewStdLogger()
)

// GetLogger returns default logger.
func GetLogger(name string) Logger {
	return With(defaultLogger, "module", name)
}

// SetLogger set a default logger.
func SetLogger(logger Logger) {
	defaultLogger = logger
}

// Logger is a logger interface.
type Logger interface {
	Print(pairs ...interface{})
}

type logger struct {
	log   Logger
	pairs []interface{}
}

func (l *logger) Print(pairs ...interface{}) {
	l.log.Print(append(pairs, l.pairs...)...)
}

// With with logger kv pairs.
func With(log Logger, pairs ...interface{}) Logger {
	if len(pairs) == 0 {
		return log
	}
	return &logger{log: log, pairs: pairs}
}

// Debug returns a debug logger.
func Debug(log Logger) Logger {
	return With(log, LevelKey, LevelDebug)
}

// Info returns a info logger.
func Info(log Logger) Logger {
	return With(log, LevelKey, LevelInfo)
}

// Warn return a warn logger.
func Warn(log Logger) Logger {
	return With(log, LevelKey, LevelWarn)
}

// Error returns a error logger.
func Error(log Logger) Logger {
	return With(log, LevelKey, LevelError)
}
