// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"path"
	"testing"

	"github.com/rs/zerolog"
)

//======================================================================================================================
// Public Functions
//======================================================================================================================

func TestExecuteCmd(t *testing.T) {
	var env = []string{"ENV1=ENV1", "ENV2=ENV2", "ENV3=ENV3"}
	var buffer LogBuffer

	// capture log output
	InitLoggerWithWriter(LogFormat(Default), &buffer, true)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	buffer = LogBuffer{}

	// test the cmd invocation
	path := path.Join(SourcePath(), "testcmd.sh")
	if err := ExecuteCmd(env, true, path, "arg1", "arg2", "arg3"); err != nil {
		t.Errorf("ExecuteCmd returned an error: %s.", err.Error())
		return
	}

	// validate at least 6 logs are returned (3 args, 3 env variables, and default env variables)
	if len(buffer) < 6 {
		t.Errorf("ExecuteCmd returned incorrect number of log messages, got: %d, want: >=6.", len(buffer))
		return
	}

	// validate arguments
	for i := 0; i < 3; i++ {
		want := fmt.Sprintf("ARG arg%d", i+1)
		if buffer[i] != want {
			t.Errorf("ExecuteCmd returned incorrect argument, got: %s, want: %s.", buffer[i], want)
		}
	}

	// validate env
	for i := 1; i <= len(env); i++ {
		want := fmt.Sprintf(`export ENV%d="ENV%d"`, i, i)
		if !Contains(buffer, want) {
			t.Errorf("ExecuteCmd did not return expected environment variable: %s.", want)
		}
	}
}
