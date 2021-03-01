// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// snapshotsCmd represents the snapshots command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Test the repository for errors",
	Long: `
The "check" command tests the repository for errors and reports any errors it
finds. By default, the "check" command will always load all data directly from the
repository and not use a local cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error {
			r, err := lib.NewResticManager()
			if err != nil {
				return err
			}
			return r.Check()
		}
		lib.HandleCmd(f, "Error executing check", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init registers the snapshotsCmd with the rootCmd, which is managed by Cobra.
func init() {
	rootCmd.AddCommand(checkCmd)
}
