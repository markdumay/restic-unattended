// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
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
		f := func() error { return Restore(RestorePath, Snapshot) }
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

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Restore retrieves a specific restic snapshot and restores it at the specified path.
func Restore(path string, snapshot string) error {
	lib.Logger.Info().Msgf("Starting restore operation for snapshot '%s'", snapshot)

	// check if the repository is already initialized, fail if not available
	if err := lib.ExecuteResticCmd(false, "snapshots"); err != nil {
		return &lib.ResticError{Err: "Could not open repository", Fatal: true}
	}

	// ensure the repository is unlocked
	if err := lib.ExecuteResticCmd(false, "unlock"); err != nil {
		return &lib.ResticError{Err: "Could not unlock repository", Fatal: true}
	}

	if err := lib.ExecuteResticCmd(true, "restore", snapshot, "--target="+path); err != nil {
		return &lib.ResticError{Err: fmt.Sprintf("Could not restore snapshot '%s'", snapshot), Fatal: true}
	}

	lib.Logger.Info().Msgf("Finished restore operation for snapshot '%s'", snapshot)
	return nil
}
