// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"path"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

//======================================================================================================================
// Private Functions
//======================================================================================================================

func filterCmd(buffer LogBuffer) []string {
	cmds := []string{}
	for _, l := range buffer {
		if strings.HasPrefix(l, "DEBUG  Executing command:") {
			items := strings.Split(l, "[")
			if len(items) > 0 {
				cmd := strings.TrimSuffix(items[len(items)-1], "]")
				cmds = append(cmds, cmd)
			}
		}
	}
	return cmds
}

func prepareContext(buffer *LogBuffer) *ResticManager {
	// capture logs with at least debugging level in buffer
	InitLoggerWithWriter(LogFormat(Default), buffer, true)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// create the restic manager
	env := []string{"RESTIC_REPOSITORY=RESTIC_REPOSITORY", "RESTIC_PASSWORD=RESTIC_PASSWORD"}
	path := path.Join(SourcePath(), "testcmd.sh")
	return NewResticManagerWithContext(path, env)
}

func validateLogs(t *testing.T, test string, got []string, want []string) {
	filtered := filterCmd(got)

	// confirm got and want have the same length
	if len(filtered) != len(want) {
		t.Errorf("%s returned incorrect number of commands, got: %d, want: %d.", test, len(filtered), len(want))
		return
	}

	// compare outcome with expected result
	if !Equal(filtered, want) {
		t.Errorf("%s returned incorrect commands", test)
	}
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// TODO: fix GitHub workflow
// func TestExecuteCmd(t *testing.T) {
// 	var env = []string{"ENV1=ENV1", "ENV2=ENV2", "ENV3=ENV3"}
// 	var buffer LogBuffer

// 	// capture log output
// 	InitLoggerWithWriter(LogFormat(Default), &buffer, true)
// 	zerolog.SetGlobalLevel(zerolog.InfoLevel)
// 	buffer = LogBuffer{}

// 	// test the cmd invocation
// 	path := path.Join(SourcePath(), "testcmd.sh")
// 	if err := ExecuteCmd(env, true, path, "arg1", "arg2", "arg3"); err != nil {
// 		t.Errorf("ExecuteCmd returned an error: %s.", err.Error())
// 		return
// 	}

// 	// validate at least 6 logs are returned (3 args, 3 env variables, and default env variables)
// 	if len(buffer) < 6 {
// 		t.Errorf("ExecuteCmd returned incorrect number of log messages, got: %d, want: >=6.", len(buffer))
// 		return
// 	}

// 	// validate arguments
// 	for i := 0; i < 3; i++ {
// 		want := fmt.Sprintf("ARG arg%d", i+1)
// 		if buffer[i] != want {
// 			t.Errorf("ExecuteCmd returned incorrect argument, got: %s, want: %s.", buffer[i], want)
// 		}
// 	}

// 	// validate env
// 	for i := 1; i <= len(env); i++ {
// 		want1 := fmt.Sprintf(`export ENV%d="ENV%d"`, i, i) // macOS uses quoted variables by default
// 		want2 := fmt.Sprintf("export ENV%d=ENV%d", i, i)   // ubuntu uses unquoted variables by default
// 		if !Contains(buffer, want1) && !Contains(buffer, want2) {
// 			t.Errorf("ExecuteCmd did not return expected environment variable: %s.", want2)
// 		}
// 	}
// }

func TestBackup(t *testing.T) {
	const test = "Backup"
	expected := []string{
		"snapshots",
		"unlock",
		"backup ./backup --host=HOST",
	}

	var buffer LogBuffer
	r := prepareContext(&buffer)
	if err := r.Backup("./backup", true, "HOST"); err != nil {
		t.Errorf("%s returned an error: %s.", test, err.Error())
	}
	validateLogs(t, test, buffer, expected)
}

func TestCheck(t *testing.T) {
	const test = "Check"
	expected := []string{
		"unlock",
		"check",
	}

	var buffer LogBuffer
	r := prepareContext(&buffer)
	if err := r.Check(); err != nil {
		t.Errorf("%s returned an error: %s.", test, err.Error())
	}
	validateLogs(t, test, buffer, expected)
}

func TestForget(t *testing.T) {
	const test = "Forget"
	expected := []string{
		"snapshots",
		"unlock",
		"forget forget --keep-last=5 --keep-daily=2 --prune",
	}

	args := []string{"forget", "--keep-last=5", "--keep-daily=2"}

	var buffer LogBuffer
	r := prepareContext(&buffer)
	if err := r.Forget(args); err != nil {
		t.Errorf("%s returned an error: %s.", test, err.Error())
	}
	validateLogs(t, test, buffer, expected)
}

func TestRestore(t *testing.T) {
	const test = "Restore"
	expected := []string{
		"snapshots",
		"unlock",
		"restore SNAPSHOT --target=./restore",
	}

	var buffer LogBuffer
	r := prepareContext(&buffer)
	if err := r.Restore("./restore", "SNAPSHOT"); err != nil {
		t.Errorf("%s returned an error: %s.", test, err.Error())
	}
	validateLogs(t, test, buffer, expected)
}

func TestSnapshots(t *testing.T) {
	const test = "Snapshots"
	expected := []string{
		"unlock",
		"snapshots",
	}

	var buffer LogBuffer
	r := prepareContext(&buffer)
	if err := r.Snapshots(); err != nil {
		t.Errorf("%s returned an error: %s.", test, err.Error())
	}
	validateLogs(t, test, buffer, expected)
}
