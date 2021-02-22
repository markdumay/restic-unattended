// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"errors"
	"os/exec"

	"github.com/rs/zerolog"
)

// ResticError defines a custom error for failed execution of restic commands.
type ResticError struct {
	Err   string // error description
	Fatal bool   // fatal or non-fatal error
}

func (e *ResticError) Error() string {
	return e.Err
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// ExecuteCmd invokes an external command with the provided arguments and enviornment variables. Pending if log is true,
// all output of the command (both stdout and stderr) is logged in real time. Otherwise, only errors are logged.
func ExecuteCmd(env []string, log bool, command string, args ...string) error {
	// initiate the restic subcommand with current environment and secrets
	cmd := exec.Command(command, args...)
	cmd.Env = env

	// redirect stdout and stderr to the default logger if instructed
	if log {
		cmd.Stdout = NewLogWriter(&Logger, zerolog.InfoLevel)
	}
	cmd.Stderr = NewLogWriter(&Logger, zerolog.ErrorLevel)

	// start the command and wait for it to finish
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

// ExecuteResticCmd invokes the external binary restic with a specific subcommand. It stages any Docker secrets as
// environment variables first. The output of the command (both stdout and stderr) is logged in real time. See
// ExecuteCmd for more details.
func ExecuteResticCmd(log bool, subCmd string, args ...string) error {
	// initialize the Docker secrets
	env, err := StageEnv()
	if err != nil {
		return err
	}

	// initiate the restic command with current environment and secrets
	resticArgs := []string{subCmd}
	resticArgs = append(resticArgs, args...)
	return ExecuteCmd(env, log, "restic", resticArgs...)
}

// HandleCmd invokes a function and handles the resulting error, if any. An error is written to the general logger
// if cmd returns an error. If the provided cmd returns an error of type lib.ResticError and is flagged to be Fatal,
// the logger receives a fatal error (and exits the program accordingly).
func HandleCmd(cmd func() error, errMsg string, alwaysFatal bool) {
	if err := cmd(); err != nil {
		var resticError *ResticError
		if alwaysFatal || (errors.As(err, &resticError) && resticError.Fatal) {
			Logger.Fatal().Err(err).Msg(errMsg)
		} else {
			Logger.Error().Err(err).Msg(errMsg)
		}
	}
}
