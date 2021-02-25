// Copyright © 2021 Mark Dumay. All rights reserved.
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

//======================================================================================================================
// Public Functions
//======================================================================================================================

func TestInitLogger(t *testing.T) {
	const layout = "2006-01-02T15:04:05Z07:00"
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

	// test Default logging for levels >= Error
	InitLoggerWithWriter(LogFormat(Default), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	buffer = LogBuffer{}
	for _, item := range table {
		// validate each log message conforms to template "ERROR  Error message"
		Logger.WithLevel(item.level).Msg(item.message)
		if item.level >= zerolog.ErrorLevel {
			log := buffer[len(buffer)-1]
			level := strings.ToUpper(item.level.String())
			msg := strings.TrimSpace(strings.TrimPrefix(log, level))

			if msg != item.message {
				t.Errorf("InitLogger with '%s' formatting returned incorrect log message, got: '%s', want: '%s'.", "Default",
					msg, item.message)
			}
		}
	}

	// test Pretty logging for all levels
	InitLoggerWithWriter(LogFormat(Pretty), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	buffer = LogBuffer{}
	for _, item := range table {
		// validate each log message conforms to template "2021-01-01T00:00:00Z | TRACE  | Trace message"
		Logger.WithLevel(item.level).Msg(item.message)
		log := buffer[len(buffer)-1]
		elements := strings.Split(string(log), "|")
		if len(elements) != 3 {
			t.Errorf("InitLogger with '%s' formatting returned incorrect number of elements, got: %d, want: %d.",
				"Pretty", len(elements), 3)
		} else {
			timestamp := strings.TrimSpace(elements[0])
			level := strings.TrimSpace(elements[1])
			msg := strings.TrimSpace(elements[2])
			// validate first element is a timestamp
			if _, err := time.Parse(layout, timestamp); err != nil {
				t.Errorf("InitLogger with '%s' formatting returned incorrect timestamp format, got: %s, want: %s.", "Pretty",
					timestamp, layout)
			}
			// validate second element has correct logging level
			if level != strings.ToUpper(item.level.String()) {
				t.Errorf("InitLogger with '%s' formatting returned incorrect level, got: %s, want: %s.", "Pretty",
					level, strings.ToUpper(item.level.String()))
			}
			// validate third element has expected log message
			if msg != item.message {
				t.Errorf("InitLogger with '%s' formatting returned incorrect message, got: %s, want: %s.", "Pretty",
					msg, item.message)
			}
		}
	}

	// test JSON logging for all levels
	InitLoggerWithWriter(LogFormat(JSON), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	buffer = LogBuffer{}
	for _, item := range table {
		// validate each log message conforms to expected JSON response
		Logger.WithLevel(item.level).Msg(item.message)
		log := buffer[len(buffer)-1]
		// validate JSON response string can be unmarshalled
		jsonLog, err := UnmarshalLog([]byte(log))
		if err != nil {
			// if err := json.Unmarshal([]byte(log), &jsonLog); err != nil {
			t.Errorf("InitLogger with '%s' formatting returned incorrect log message, error: '%s'.", "JSON",
				err.Error())
			continue
		}

		// validate first element is a timestamp and within expected range
		currentTime := time.Now()
		diff := currentTime.Sub(jsonLog.Time)
		if diff.Minutes() > 1 {
			t.Errorf("InitLogger with '%s' formatting returned incorrect timestamp, got: %s, want: %s (+/- 1 min).", "JSON",
				jsonLog.Time.Format(layout), currentTime.Format(layout))
		}
		// validate logging level
		if jsonLog.Level != item.level {
			t.Errorf("InitLogger with '%s' formatting returned incorrect level, got: %d, want: %d (%s).", "JSON",
				jsonLog.Level, item.level, strings.ToUpper(item.level.String()))
		}
		// validate log message
		if jsonLog.Message != item.message {
			t.Errorf("InitLogger with '%s' formatting returned incorrect message, got: %s, want: %s.", "Pretty",
				jsonLog.Message, item.message)
		}
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
