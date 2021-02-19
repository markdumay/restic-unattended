// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"errors"

	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// BackupCron defines the schedule for the backup cron job. See https://github.com/robfig/cron/ for the cron spec
// format.
var BackupCron string

// ForgetCron defines the schedule for the forget cron job, similar to BackupCron.
var ForgetCron string

// Sustained defines if processing of scheduled jobs should continue despite errors
var Sustained bool

// scheduleCmd represents the schedule command. It sets up a job that is repeated following a cron schedule. It requires
// one argument that represents the cron spec.
var scheduleCmd = &cobra.Command{
	Use:   "schedule <cron>",
	Short: "Run a backup using cron schedule",
	Long: `
Schedule sets up a backup job that is repeated following a cron schedule. It
optionally removes old snapshots using a policy too. The cron notation supports
optional seconds. The following expressions are supported:
Field name   | Mandatory? | Allowed values  | Allowed special characters
----------   | ---------- | --------------  | --------------------------
Seconds      | No         | 0-59            | * / , -
Minutes      | Yes        | 0-59            | * / , -
Hours        | Yes        | 0-23            | * / , -
Day of month | Yes        | 1-31            | * / , - ?
Month        | Yes        | 1-12 or JAN-DEC | * / , -
Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

Special characters:
Asterisk ( * )
The asterisk indicates that the cron expression will match for all values of 
the field; e.g., using an asterisk in the 5th field (month) would indicate every 
month.

Slash ( / )
Slashes are used to describe increments of ranges. For example 3-59/15 in the 
1st field (minutes) would indicate the 3rd minute of the hour and every 15 
minutes thereafter. The form "*\/..." is equivalent to the form 
"first-last/...", that is, an increment over the largest possible range of the 
field. The form "N/..." is accepted as meaning "N-MAX/...", that is, starting 
at N, use the increment until the end of that specific range. It does not wrap 
around.

Comma ( , )
Commas are used to separate items of a list. For example, using "MON,WED,FRI"
in the 5th field (day of week) would mean Mondays, Wednesdays, and Fridays.

Hyphen ( - )
Hyphens are used to define ranges. For example, 9-17 would indicate every hour
between 9am and 5pm inclusive.

Question mark ( ? )
Question mark may be used instead of '*' for leaving either day-of-month or 
day-of-week blank.

Predefined schedules:
The following predefined schedules may be used instead of the common cron 
fields:
@yearly (or @annually), @monthly, @weekly, @daily (or @midnight), and @hourly.

Examples:
restic-unattended schedule '0 0,12 * * *'
Runs a scheduled backup at midnight and noon every day.

restic-unattended schedule '30 5 * * * *' --forget '0 0 * * *' --keep-daily 7
Runs a scheduled backup at minute 5 and 30 seconds of every hour. Recent 
backups are kept for each of the last 7 days, determined at 00:00 every day.

restic-unattended schedule '@weekly'
Runs a scheduled backup once a week at midnight on Sunday.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a cron argument")
		}
		if err := lib.IsValidCron(args[0]); err != nil {
			return err
		}
		BackupCron = args[0]
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		initScheduleFlags(cmd.Flags())
		return validateScheduleFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error { return Schedule(BackupPath, InitRepository, Host, Sustained, cmd.Flags()) }
		lib.HandleCmd(f, "Error running schedule command", true)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init registers the scheduleCmd with the rootCmd, which is managed by Cobra.
func init() {
	scheduleCmd.Flags().StringVar(&ForgetCron, "forget", "", "remove old snapshots according to rotation schedule.")
	scheduleCmd.Flags().BoolVar(&Sustained, "sustained", false, "sustain processing of scheduled jobs despite errors")

	addBackupOptions(scheduleCmd)
	addKeepOptions(scheduleCmd)
	rootCmd.AddCommand(scheduleCmd)
}

// initScheduleFlags validates the provided persistent flags and initializes applicable global values. Currently
// supported flag is "logformat". By default, logs are printed using pretty formatting, unless explicitly set to
// another log format.
func initScheduleFlags(flags *pflag.FlagSet) {
	if !viper.IsSet("logformat") {
		lib.InitLogger(lib.LogFormat(lib.Pretty))
	}
}

func validateScheduleFlags(flags *pflag.FlagSet) error {
	if BackupPath == "" {
		return errors.New("No backup path provided")
	}

	if ForgetCron != "" {
		return lib.IsValidCron(ForgetCron)
	}

	return nil
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Schedule starts the cron job following the provided BackupCron. If needed, the repository is initialized first. The
// cron job runs indefinitely, unless interrupted (e.g. pressing Ctrl-C or sending SIGINT).
func Schedule(path string, init bool, host string, sustain bool, keepFlags *pflag.FlagSet) error {
	lib.Logger.Info().Msg("Executing schedule command")

	var jobs []lib.Job

	if BackupCron != "" {
		var backup lib.Job
		backup.Tag = "backup"
		backup.Spec = BackupCron
		backup.RunE = func() error { return Backup(path, init, host) }
		jobs = append(jobs, backup)
	}

	if ForgetCron != "" {
		var forget lib.Job
		forget.Tag = "forget"
		forget.Spec = ForgetCron
		forget.RunE = func() error { return Forget(keepFlags) }
		jobs = append(jobs, forget)
	}

	return lib.RunCronJobs(jobs, !sustain)
}
