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
var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "List all snapshots",
	Long: `
The "snapshots" command lists all snapshots stored in the repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error {
			r := lib.NewResticManager()
			return r.Snapshots()
		}
		lib.HandleCmd(f, "Error retrieving snapshots", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init registers the snapshotsCmd with the rootCmd, which is managed by Cobra.
func init() {
	rootCmd.AddCommand(snapshotsCmd)
}
