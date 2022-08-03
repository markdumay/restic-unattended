// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"os"

	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// RestorePath specifies the target path to restore the files to.
var RestorePath string

// Snapshot defines the snapshot ID to restore from (defaults to "latest")
var Snapshot string = "latest"

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore <path>",
	Short: "Restore a remote backup to a local path",
	Long: `
Restores a backup stored in a restic repository to a local path.
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a path argument")
		}
		if _, err := os.Stat(args[0]); err != nil {
			return err
		}
		RestorePath = args[0]
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error {
			r, err := lib.NewResticManager()
			if err != nil {
				return err
			}
			return r.Restore(RestorePath, Snapshot)
		}
		lib.HandleCmd(f, "Error running restore", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init registers the restoreCmd with the rootCmd, which is managed by Cobra.
func init() {
	restoreCmd.Flags().StringVarP(&Snapshot, "snapshot", "", "latest",
		"ID of the snapshot to restore")
	rootCmd.AddCommand(restoreCmd)
}
