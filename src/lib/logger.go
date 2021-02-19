package lib

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

//======================================================================================================================
// Variables and user-defined types
//======================================================================================================================

// StructuredLog instructs the logger to print structured logs to the console for easier log aggregation. It is used by
// the schedule command by default.
var structuredLog bool = false

// logFormat defines the output format of the logger. It supports three types of formatting. The schedule command uses
// Pretty formatting by default, the other commands use Default formatting. The formatting can be specified using the
// environment variable RESTIC_LOGFORMAT or the global flag --logformat.
//  - Default prints logs as standard console output (no timestamp and level prefixes)
//  - Pretty prints logs as semi-structured messages with a timestamp and level prefix
//  - JSON prints logs as JSON strings
var logFormat LogFormat = LogFormat(Default)

// Logger is the global logger.
var Logger zerolog.Logger

// LogWriter is a user-defined type for a custom writer.
type LogWriter struct {
	logger *zerolog.Logger
	level  zerolog.Level
}

// LogFormat defines the type of logging format to use.
type LogFormat int

// Defines a pseudo enumeration of possible logging formats.
const (
	// Default prints logs as standard console output (no timestamp and level prefixes), for example:
	// > Listing snapshots
	Default int = iota
	// Pretty prints logs as semi-structured messages with a timestamp and level prefix, for example:
	// > 2020-12-17T07:12:57+01:00 | INFO   | Listing snapshots
	Pretty
	// JSON prints logs as JSON strings, for example:
	// > {"level":"info","time":"2020-12-17T07:12:57+01:00","message":"Listing snapshots"}
	JSON
)

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init calls InitLogger to initialize the default logger.
func init() {
	InitLogger(logFormat)
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// InitLogger initializes the global logger. If structured is set, the logger is instructed to print structured logs
// to the console for easier log aggregation. It is used by the schedule command by default.
func InitLogger(format LogFormat) {
	logFormat = format
	var output io.Writer

	switch logFormat {
	case LogFormat(Default):
		writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		writer.FormatTimestamp = func(i interface{}) string {
			return ""
		}
		writer.FormatLevel = func(i interface{}) string {
			v, ok := i.(string)
			if ok && v == "info" {
				return ""
			}
			return strings.ToUpper(fmt.Sprintf("%-6s", i))
		}
		output = writer
	case LogFormat(Pretty):
		writer := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		writer.FormatTimestamp = nil
		writer.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s |", i))
		}
		output = writer
	case LogFormat(JSON):
		output = os.Stdout
	}

	Logger = zerolog.New(output).With().Timestamp().Logger()
}

// NewLogWriter returns the global logger with an instruction to write a message at the specified level.
func NewLogWriter(l *zerolog.Logger, level zerolog.Level) *LogWriter {
	lw := &LogWriter{}
	lw.logger = l
	lw.level = level
	return lw
}

// Write implements the io.Writer interface for a LogWriter.
func (lw LogWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		if line != "" || !structuredLog {
			Logger.WithLevel(lw.level).Msg(line)
		}
	}
	return len(p), nil
}

// ParseFormat converts a format string into a typed logformat value.
// returns an error if the input string does not match known values.
func ParseFormat(formatStr string) (LogFormat, error) {
	switch formatStr {
	case "default":
		return LogFormat(Default), nil
	case "pretty":
		return LogFormat(Pretty), nil
	case "json":
		return LogFormat(JSON), nil
	}
	return LogFormat(Default), fmt.Errorf("Unknown Log Format String: '%s', using default", formatStr)
}
