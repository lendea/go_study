package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

// Logger is a simplified abstraction of the zap.Logger
type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Infof(template string, args ...interface{})
	Error(msg string, fields ...zapcore.Field)
	Errorf(template string, args ...interface{})
	Warn(msg string, fields ...zapcore.Field)
	Warnf(msg string, fields ...interface{})
	Fatal(msg string, fields ...zapcore.Field)
	With(fields ...zapcore.Field) Logger
	SetTag(fields ...interface{})
}

// logger delegates all calls to the underlying zap.Logger
type logger struct {
	logger *zap.Logger
	ctx    context.Context
}

var _ Logger = &logger{}

// Info logs an info msg with fields
func (l logger) Info(msg string, fields ...zapcore.Field) {
	l.setName().Info(msg, fields...)
}

func (l logger) Infof(template string, args ...interface{}) {
	l.setName().Sugar().Infof(template, args...)
}

// Error logs an error msg with fields
func (l logger) Error(msg string, fields ...zapcore.Field) {
	l.setName().Error(msg, fields...)
}

func (l logger) Errorf(template string, args ...interface{}) {
	l.setName().Sugar().Errorf(template, args...)
}

// Error logs an warn msg with fields
func (l logger) Warn(msg string, fields ...zapcore.Field) {
	l.setName().Warn(msg, fields...)
}

func (l logger) Warnf(template string, args ...interface{}) {
	l.setName().Sugar().Warnf(template, args...)
}

// Fatal logs a fatal error msg with fields
func (l logger) Fatal(msg string, fields ...zapcore.Field) {
	l.setName().Fatal(msg, fields...)
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l logger) With(fields ...zapcore.Field) Logger {
	return logger{logger: l.logger.With(fields...)}
}

func (l logger) SetTag(...interface{}) {
}

func (l logger) setName() *zap.Logger {
	return setServiceLoggerName(l.logger)
}

// setServiceLoggerName
func setServiceLoggerName(l *zap.Logger) *zap.Logger {
	name := gServiceName
	return l.Named(name)
}

////////////////// wrap as io.Writer //////////////////////
type LoggerInfoWriter interface {
	io.Writer
}

func BuildInfoWriter(l Logger) LoggerInfoWriter {
	return loggerInfoWriter{l}
}

type loggerInfoWriter struct {
	Logger
}

func (w loggerInfoWriter) Write(p []byte) (n int, err error) {
	w.Info(string(p))
	return len(p), nil
}

type LoggerErrorWriter interface {
	io.Writer
}

func BuildErrorWriter(l Logger) LoggerErrorWriter {
	return loggerErrorWriter{l}
}

type loggerErrorWriter struct {
	Logger
}

func (w loggerErrorWriter) Write(p []byte) (n int, err error) {
	w.Error(string(p))
	return len(p), nil
}
