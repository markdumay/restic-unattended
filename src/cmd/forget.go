// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
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
		f := func() error {
			r, err := lib.NewResticManager()
			if err != nil {
				return err
			}
			args, err := lib.ParseArgs(cmd.Flags(), "^keep-")
			if err != nil {
				return err
			}
			return r.Forget(args)
		}
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
