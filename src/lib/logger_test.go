// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type LogBuffer []string

//======================================================================================================================
// Private Functions
//======================================================================================================================

func isValidLevel(t *testing.T, input interface{}, expected zerolog.Level, loggingType string) bool {
	var level zerolog.Level

	// convert input to typed level, fail on error
	switch v := input.(type) {
	case string:
		s := strings.ToLower(strings.TrimSpace(input.(string)))
		parsed, err := zerolog.ParseLevel(s)
		if err != nil {
			t.Errorf("InitLogger with '%s' formatting could not convert log level", loggingType)
			return false
		}
		level = parsed
	case zerolog.Level:
		parsed, ok := input.(zerolog.Level)
		if !ok {
			t.Errorf("InitLogger with '%s' formatting could not convert log level", loggingType)
			return false
		}
		level = parsed
	default:
		t.Errorf("InitLogger with '%s' formatting could not convert log level of type %T", loggingType, v)
		return false
	}

	// validate level equals expected value
	if level != expected {
		t.Errorf("InitLogger with '%s' formatting returned incorrect level, got: %s, want: %s.", loggingType,
			level.String(), expected.String())
	}

	return true
}

func isValidMessage(t *testing.T, input string, expected string, loggingType string) bool {
	s := strings.TrimSpace(input)
	if s != expected {
		t.Errorf("InitLogger with '%s' formatting returned incorrect message, got: %s, want: %s.", loggingType,
			s, expected)
		return false
	}

	return true
}

func isValidTimestamp(t *testing.T, input interface{}, expected time.Time, loggingType string) bool {
	const layout = "2006-01-02T15:04:05Z07:00"
	var timestamp time.Time

	// convert input to typed timestamp, fail on error
	switch v := input.(type) {
	case string:
		s, ok := input.(string)
		if !ok {
			t.Errorf("InitLogger with '%s' formatting could not convert timestamp string", loggingType)
			return false
		}
		time, err := time.Parse(layout, strings.TrimSpace(s))
		if err != nil {
			t.Errorf("InitLogger with '%s' formatting returned incorrect timestamp format, got: %s, want: %s.", "Pretty",
				timestamp, layout)
			return false
		}
		timestamp = time
	case time.Time:
		time, ok := input.(time.Time)
		if !ok {
			t.Errorf("InitLogger with '%s' formatting could not convert timestamp", loggingType)
			return false
		}
		timestamp = time
	default:
		t.Errorf("InitLogger with '%s' formatting could not convert timestamp of type %T", loggingType, v)
		return false
	}

	// validate timestamp is within expected range
	diff := expected.Sub(timestamp)
	if diff.Minutes() > 1 {
		t.Errorf("InitLogger with '%s' formatting returned incorrect timestamp, got: %s, want: %s (+/- 1 min).", "JSON",
			timestamp, expected.Format(time.RFC3339))
		return false
	}

	return true
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

func TestInitLogger(t *testing.T) {
	var buffer LogBuffer
	table := []struct {
		level   zerolog.Level
		message string
	}{
		{zerolog.TraceLevel, "Trace message"},
		{zerolog.DebugLevel, "Debug message"},
		{zerolog.InfoLevel, "Info message"},
		{zerolog.WarnLevel, "Warn message"},
		{zerolog.ErrorLevel, "Error message"},
		{zerolog.FatalLevel, "Fatal message"},
		{zerolog.PanicLevel, "Panic message"},
	}

	// test Default logs for levels >= Error conforms to template "ERROR  Error message"
	InitLoggerWithWriter(LogFormat(Default), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	buffer = LogBuffer{}
	for _, item := range table {
		// generate and capture a log message
		Logger.WithLevel(item.level).Msg(item.message)
		if item.level < zerolog.ErrorLevel {
			continue
		}
		log := buffer[len(buffer)-1]

		// validate level and message (level is tested implicitly)
		level := strings.ToUpper(item.level.String())
		msg := strings.TrimPrefix(log, level)
		isValidMessage(t, msg, item.message, "Default")
	}

	// test Pretty logs for all levels conforms to template "2021-01-01T00:00:00Z | TRACE  | Trace message"
	InitLoggerWithWriter(LogFormat(Pretty), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	buffer = LogBuffer{}
	for _, item := range table {
		// generate and capture a log message
		Logger.WithLevel(item.level).Msg(item.message)
		log := buffer[len(buffer)-1]
		elements := strings.Split(string(log), "|")
		if len(elements) != 3 {
			t.Errorf("InitLogger with '%s' formatting returned incorrect number of elements, got: %d, want: %d.",
				"Pretty", len(elements), 3)
			continue
		}

		// validate timestamp, level, and message
		isValidTimestamp(t, elements[0], time.Now(), "Pretty")
		isValidLevel(t, elements[1], item.level, "Pretty")
		isValidMessage(t, elements[2], item.message, "Pretty")
	}

	// test JSON logs for all levels
	InitLoggerWithWriter(LogFormat(JSON), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	buffer = LogBuffer{}
	for _, item := range table {
		// generate and capture a log message
		Logger.WithLevel(item.level).Msg(item.message)
		log := buffer[len(buffer)-1]

		// validate that JSON response string can be unmarshalled
		jsonLog, err := UnmarshalLog([]byte(log))
		if err != nil {
			t.Errorf("InitLogger with '%s' formatting returned incorrect log message, error: '%s'.", "JSON",
				err.Error())
			continue
		}

		// validate timestamp, level, and message
		isValidTimestamp(t, jsonLog.Time, time.Now(), "Pretty")
		isValidLevel(t, jsonLog.Level, item.level, "Pretty")
		isValidMessage(t, jsonLog.Message, item.message, "Pretty")
	}
}

// Write implements the io.Writer interface for a simple in-memory buffer. Lines are separated by newline characters
// and added one by one.
func (b *LogBuffer) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		if line != "" {
			*b = append(*b, line)
		}
	}
	return len(p), nil
}
