package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

const (
	levelFatal    = slog.Level(12)
	fatalExitCode = 1
)

var levelNames = map[slog.Leveler]string{
	levelFatal: "FATAL",
}

type (
	// ContextFieldExtractor is a function that receives a context and should
	// return a slice of loggerApi.Attribute to be added into every log call.
	ContextFieldExtractor func(ctx context.Context) []Attribute
)

type Logger struct {
	logger         *slog.Logger
	errorLogger    *slog.Logger
	level          *logLeveler
	fieldExtractor ContextFieldExtractor
}

type Options struct {
	TextOutput            bool
	LogOnlyFatalLevel     bool
	FixedAttributes       map[string]string
	ContextFieldExtractor ContextFieldExtractor
}

// New creates a new Logger interface for applications.
func New(options Options) *Logger {
	var (
		attrs []slog.Attr
		level = newLogLeveler(slog.LevelInfo)
		opts  = &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Prints our custom log level label.
				if a.Key == slog.LevelKey {
					if level, ok := a.Value.Any().(slog.Level); ok {
						levelLabel, exists := levelNames[level]
						if !exists {
							levelLabel = level.String()
						}

						a.Value = slog.StringValue(levelLabel)
					}
				}

				// Change the source path to only 'dir/file.go'
				if a.Key == slog.SourceKey {
					if source, ok := a.Value.Any().(*slog.Source); ok {
						filename := filepath.Base(source.File)
						source.File = filepath.Join(filepath.Base(filepath.Dir(source.File)), filename)
					}
				}

				return a
			},
		}
	)

	// Adds custom fixed attributes into every log message.
	for k, v := range options.FixedAttributes {
		attrs = append(attrs, slog.String(k, v))
	}

	logHandler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs(attrs)
	if options.TextOutput {
		logHandler = slog.NewTextHandler(os.Stdout, opts).WithAttrs(attrs)
	}

	// Creates a specific log handler so every error message can have its source
	// in the output.
	opts.AddSource = true
	errHandler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs(attrs)
	if options.TextOutput {
		errHandler = slog.NewTextHandler(os.Stdout, opts).WithAttrs(attrs)
	}

	// This configures the test environment to only log fatal errors, so the
	// test output is easier to read and debug.
	if options.LogOnlyFatalLevel {
		level.setLevel(levelFatal)
	}

	return &Logger{
		logger:         slog.New(logHandler),
		errorLogger:    slog.New(errHandler),
		level:          level,
		fieldExtractor: options.ContextFieldExtractor,
	}
}

// Debug outputs messages using debug level.
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Debug(msg, mFields...)
}

// Info outputs messages using the info level.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Info(msg, mFields...)
}

// Warn outputs messages using warning level.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Warn(msg, mFields...)
}

// Error outputs messages using error level.
func (l *Logger) Error(ctx context.Context, msg string, attrs ...Attribute) {
	l.error(ctx, msg, attrs...)
}

func (l *Logger) error(ctx context.Context, msg string, attrs ...Attribute) {
	var (
		mFields = l.mergeFieldsWithCtx(ctx, attrs)
		pcs     [1]uintptr
	)

	if l.level.Level() > slog.LevelError {
		return
	}

	runtime.Callers(3, pcs[:]) // skip [Callers, error]
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])

	if len(mFields) > 0 {
		r.Add(mFields...)
	}

	_ = l.errorLogger.Handler().Handle(ctx, r)
}

// Fatal outputs message using fatal level.
func (l *Logger) Fatal(ctx context.Context, msg string, attrs ...Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Log(ctx, levelFatal, msg, mFields...)
	os.Exit(fatalExitCode)
}

func (l *Logger) mergeFieldsWithCtx(ctx context.Context, attrs []Attribute) []any {
	var (
		appendedFields = l.appendServiceContext(ctx, attrs)
		mergedFields   = make([]any, len(appendedFields))
	)

	for i, field := range appendedFields {
		mergedFields[i] = slog.Any(field.Key(), field.Value())
	}

	return mergedFields
}

// DisableDebugMessages is a helper method to disable Debug level messages.
func (l *Logger) DisableDebugMessages() {
	l.level.setLevel(slog.LevelInfo)
}

// appendServiceContext executes a custom field extractor from the current
// context to add more fields into the message.
func (l *Logger) appendServiceContext(ctx context.Context, attrs []Attribute) []Attribute {
	if l.fieldExtractor != nil {
		attrs = append(attrs, l.fieldExtractor(ctx)...)
	}

	return attrs
}

// SetLogLevel changes the current messages log level.
func (l *Logger) SetLogLevel(level string) (string, error) {
	var newLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		newLevel = slog.LevelDebug
	case "info":
		newLevel = slog.LevelInfo
	case "warn":
		newLevel = slog.LevelWarn
	case "error":
		newLevel = slog.LevelError
	case "fatal":
		newLevel = levelFatal
	default:
		return "", fmt.Errorf("unknown log level '%v'", level)
	}

	l.level.setLevel(newLevel)
	return level, nil
}

// Level gets the current log level.
func (l *Logger) Level() string {
	switch l.level.Level() {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	case levelFatal:
		return "fatal"
	}

	return "unknown"
}

func (l *Logger) Debugf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, Any(k, v))
		}
	}

	l.Debug(ctx, msg, loggerFields...)
}

func (l *Logger) Infof(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, Any(k, v))
		}
	}

	l.Info(ctx, msg, loggerFields...)
}

func (l *Logger) Warnf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, Any(k, v))
		}
	}

	l.Warn(ctx, msg, loggerFields...)
}

func (l *Logger) Errorf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, Any(k, v))
		}
	}

	l.Error(ctx, msg, loggerFields...)
}

func (l *Logger) Fatalf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, Any(k, v))
		}
	}

	l.Fatal(ctx, msg, loggerFields...)
}
