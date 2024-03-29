package lib

import (
	"encoding/json"
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

// LogMessage defines the structure of JSON-formatted log messages produced by zerolog.
type LogMessage struct {
	Level   zerolog.Level `json:"level"`
	Time    time.Time     `json:"time"`
	Message string        `json:"message"`
}

// Defines a pseudo enumeration of possible logging formats.
const (
	// Default prints logs as standard console output (no timestamp and level prefixes), for example:
	// > Listing snapshots
	Default LogFormat = iota
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

// InitLoggerWithWriter initializes the global logger with the desired format and writer.
func InitLoggerWithWriter(format LogFormat, w io.Writer, noColor bool) {
	logFormat = format
	var output io.Writer

	switch logFormat {
	case LogFormat(Default):
		writer := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339, NoColor: noColor}
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
		writer := zerolog.ConsoleWriter{Out: w, TimeFormat: time.RFC3339, NoColor: noColor}
		writer.FormatTimestamp = nil
		writer.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s |", i))
		}
		output = writer
	case LogFormat(JSON):
		output = w
	}

	Logger = zerolog.New(output).With().Timestamp().Logger()
}

// InitLogger initializes the global logger with the desired format.
func InitLogger(format LogFormat) {
	InitLoggerWithWriter(format, os.Stdout, false)
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
		// skip empty lines when not using default logging format
		if line != "" || logFormat == LogFormat(Default) {
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

// String converts a typed log format to it's string representation.
func (format LogFormat) String() string {
	return [...]string{"default", "pretty", "json"}[format]
}

// UnmarshalLog converts json bytes into a LogMessage instance.
func UnmarshalLog(bytes []byte) (*LogMessage, error) {
	const layout = "2006-01-02T15:04:05Z07:00"

	// construct a placeholder with looser typing
	raw := struct {
		Level   string `json:"level"`
		Time    string `json:"time"`
		Message string `json:"message"`
	}{}

	// convert json input to placeholder type
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}

	// convert input to typed timestamp, fail on error
	timestamp, err := time.Parse(layout, raw.Time)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse datetime format, got %s, want %s", raw.Time, layout)
	}

	// parse Level
	level, err := zerolog.ParseLevel(raw.Level)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse level: %s", raw.Level)
	}

	// convert placeholder type to final type
	log := &LogMessage{
		Level:   level,
		Time:    timestamp,
		Message: raw.Message,
	}

	return log, nil
}
