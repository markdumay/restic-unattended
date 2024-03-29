// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/rs/zerolog"
)

// ResticManager manages the invocation of the external binary restic.
type ResticManager struct {
	cmd string
	env []string
}

// ResticError defines a custom error for failed execution of restic commands.
type ResticError struct {
	Err   string // error description
	Fatal bool   // fatal or non-fatal error
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

func (e *ResticError) Error() string {
	return e.Err
}

// ExecuteCmd invokes an external command with the provided arguments and environment variables. Pending if log is true,
// all output of the command (both stdout and stderr) is logged in real time. Otherwise, only errors are logged.
func ExecuteCmd(env []string, log bool, command string, args ...string) error {
	// initiate the command with current environment and secrets
	Logger.Debug().Msgf("Executing command: %s %s", command, args)
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

// NewResticManager creates a new restic manager.
func NewResticManager() (*ResticManager, error) {
	// initialize the Docker secrets
	m := NewSecretsManager()
	env, err := m.StageEnv()
	if err != nil {
		return nil, err
	}

	return &ResticManager{cmd: "restic", env: env}, nil
}

// NewResticManagerWithContext creates a new restic manager with a specific command to invoke.
func NewResticManagerWithContext(cmd string, env []string) *ResticManager {
	return &ResticManager{cmd: cmd, env: env}
}

// Backup performs a backup of the provided backup path and stores it in a restic repository. It uses the environment
// settings defined in lib.GetSupportedSecrets and lib.GetSupportedVariables.
func (r *ResticManager) Backup(path string, init bool, host string) error {
	Logger.Info().Msgf("Starting backup operation of path '%s'", path)

	// check if the repository is already initialized and do so if instructed
	if err := r.Execute(false, "snapshots"); err != nil {
		if init {
			Logger.Info().Msg("Initializing repository for first use")
			if err := r.Execute(true, "init"); err != nil {
				return &ResticError{Err: "Could not init repository", Fatal: true}
			}
		} else {
			return &ResticError{Err: "Could not open repository", Fatal: true}
		}
	}

	// ensure the repository is unlocked
	if err := r.Execute(false, "unlock"); err != nil {
		return &ResticError{Err: "Could not unlock repository", Fatal: true}
	}

	// execute the backup command
	args := []string{path}
	if host != "" {
		args = append(args, "--host="+host)
	}
	if err := r.Execute(true, "backup", args...); err != nil {
		return err
	}

	Logger.Info().Msgf("Finished backup operation of path '%s'", path)
	return nil
}

// Check tests the repository for errors and reports any errors it finds.
func (r *ResticManager) Check() error {
	Logger.Info().Msg("Executing check")

	// ensure the repository is unlocked
	if err := r.Execute(false, "unlock"); err != nil {
		return &ResticError{Err: "Could not open repository", Fatal: true}
	}

	// execute the snapshots command
	if err := r.Execute(true, "check"); err != nil {
		return &ResticError{Err: "Could not execute check", Fatal: true}
	}

	Logger.Info().Msgf("Finished executing check")
	return nil
}

// Execute invokes an external binary with a specific subcommand. It stages any Docker secrets as environment variables
// first. The output of the command (both stdout and stderr) is logged in real time. See executeCmd for more details.
func (r *ResticManager) Execute(log bool, subCmd string, args ...string) error {
	// initiate the restic command with current environment and secrets
	resticArgs := []string{subCmd}
	resticArgs = append(resticArgs, args...)
	return ExecuteCmd(r.env, log, r.cmd, resticArgs...)
}

// Forget executes the restic forget command. The '--prune' flag is added by default. Provided keep-* flags are relayed
// to the restic binary. Any stale locks on the repository are removed first.
func (r *ResticManager) Forget(args []string) error {
	Logger.Info().Msg("Starting forget operation")

	// check if the repository is already initialized
	if err := r.Execute(false, "snapshots"); err != nil {
		return &ResticError{Err: "Could not open repository", Fatal: true}
	}

	// ensure the repository is unlocked
	if err := r.Execute(false, "unlock"); err != nil {
		return &ResticError{Err: "Could not unlock repository", Fatal: true}
	}

	// execute the forget command
	args = append(args, "--prune") // add --prune flag by default
	if err := r.Execute(true, "forget", args...); err != nil {
		return errors.New("Could not complete forget operation")
	}

	Logger.Info().Msgf("Finished forget operation")
	return nil
}

// Restore retrieves a specific restic snapshot and restores it at the specified path.
func (r *ResticManager) Restore(path string, snapshot string) error {
	Logger.Info().Msgf("Starting restore operation for snapshot '%s'", snapshot)

	// check if the repository is already initialized, fail if not available
	if err := r.Execute(false, "snapshots"); err != nil {
		return &ResticError{Err: "Could not open repository", Fatal: true}
	}

	// ensure the repository is unlocked
	if err := r.Execute(false, "unlock"); err != nil {
		return &ResticError{Err: "Could not unlock repository", Fatal: true}
	}

	if err := r.Execute(true, "restore", snapshot, "--target="+path); err != nil {
		return &ResticError{Err: fmt.Sprintf("Could not restore snapshot '%s'", snapshot), Fatal: true}
	}

	Logger.Info().Msgf("Finished restore operation for snapshot '%s'", snapshot)
	return nil
}

// Schedule starts the cron job following the provided BackupCron. If needed, the repository is initialized first. The
// cron job runs indefinitely, unless interrupted (e.g. pressing Ctrl-C or sending SIGINT).
func (r *ResticManager) Schedule(backupCron string, forgetCron, path string, init bool, host string, sustain bool,
	keepFlags []string) error {

	Logger.Info().Msg("Executing schedule command")

	var jobs []Job

	if backupCron != "" {
		var backup Job
		backup.Tag = "backup"
		backup.Spec = backupCron
		backup.RunE = func() error {
			return r.Backup(path, init, host)
		}
		jobs = append(jobs, backup)
	}

	if forgetCron != "" {
		var forget Job
		forget.Tag = "forget"
		forget.Spec = forgetCron
		forget.RunE = func() error { return r.Forget(keepFlags) }
		jobs = append(jobs, forget)
	}

	return RunCronJobs(jobs, !sustain)
}

// Snapshots lists all snapshots stored in the repository.
func (r *ResticManager) Snapshots() error {
	Logger.Info().Msg("Listing snapshots")

	// ensure the repository is unlocked
	if err := r.Execute(false, "unlock"); err != nil {
		return &ResticError{Err: "Could not open repository", Fatal: true}
	}

	// execute the snapshots command
	if err := r.Execute(true, "snapshots"); err != nil {
		return &ResticError{Err: "Could not list snapshots", Fatal: true}
	}

	Logger.Info().Msgf("Finished listing snapshots")
	return nil
}
