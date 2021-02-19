// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"regexp"

	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// forgetCmd represents the forget command
var forgetCmd = &cobra.Command{
	Use:   "forget",
	Short: "Remove old snapshots according to rotation schedule",
	Long: `
Forget removes old backups according to a rotation schedule. It both flags 
snapshots for removal as well as deletes (prunes) the actual old snapshot from
the repository.

Examples:
restic-unattended forget --keep-last 5
Keep the 5 most recent snapshots

restic-unattended forget --keep-daily 7
Keep the most recent backup for each of the last 7 days
`,
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error { return Forget(cmd.Flags()) }
		lib.HandleCmd(f, "Error running forget", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

func addKeepOptions(c *cobra.Command) {
	f := c.Flags()
	f.Int("keep-last", 0, "never delete the n last (most recent) snapshots")
	f.Int("keep-hourly", 0,
		"for the last n hours in which a snapshot was made, keep only the last snapshot for each hour")
	f.Int("keep-daily", 0,
		"for the last n days which have one or more snapshots, only keep the last one for that day")
	f.Int("keep-weekly", 0,
		"for the last n weeks which have one or more snapshots, only keep the last one for that week")
	f.Int("keep-monthly", 0,
		"for the last n months which have one or more snapshots, only keep the last one for that month")
	f.Int("keep-yearly", 0,
		"for the last n years which have one or more snapshots, only keep the last one for that year")
	f.StringArray("keep-tag", []string{},
		"keep all snapshots which have all tags specified by this option (can be specified multiple times)")
	f.String("keep-within", "",
		"keep all snapshots which have been made within the duration of the latest snapshot")
}

// init registers the forgetCmd with the rootCmd, which is managed by Cobra. It adds several "keep-*" flags to define
// the exact backup rotation schedule.
func init() {
	addKeepOptions(forgetCmd)
	forgetCmd.Flags().SortFlags = false
	rootCmd.AddCommand(forgetCmd)
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Forget executes the restic forget command. The '--prune' flag is added by default. Provided keep-* flags are relayed
// to the restic binary. Any stale locks on the repository are removed first.
func Forget(flags *pflag.FlagSet) error {
	lib.Logger.Info().Msg("Starting forget operation")

	// prepare forget args
	var args = []string{"--prune"} // add --prune flag by default
	re, err := regexp.Compile("^keep-")
	if err != nil {
		return &lib.ResticError{Err: "Could not parse forget arguments", Fatal: true}
	}

	var parseErr error
	flags.Visit(func(flag *pflag.Flag) {
		// stop processing additional flags if there was an error
		if parseErr != nil {
			return
		}
		// process keep-* flags
		if re.MatchString(flag.Name) {
			v, err := lib.GetCLIFlag(flags, flag)
			if err != nil {
				parseErr = err
				return
			}
			args = append(args, v...)
		}
	})
	if parseErr != nil {
		return &lib.ResticError{Err: "Could not parse forget arguments", Fatal: true}
	}

	// check if the repository is already initialized
	if err := lib.ExecuteResticCmd(false, "snapshots"); err != nil {
		return &lib.ResticError{Err: "Could not open repository", Fatal: true}
	}

	// ensure the repository is unlocked
	if err := lib.ExecuteResticCmd(false, "unlock"); err != nil {
		return &lib.ResticError{Err: "Could not unlock repository", Fatal: true}
	}

	// execute the forget command
	if err := lib.ExecuteResticCmd(true, "forget", args...); err != nil {
		return errors.New("Could not complete forget operation")
	}

	lib.Logger.Info().Msgf("Finished forget operation")
	return nil
}
