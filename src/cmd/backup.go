// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"errors"

	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// BackupPath specifies the source path to backup
var BackupPath string

// Host to use in backups (defaults to $HOSTNAME)
var Host string

// InitRepository initializes the repository if it does not exist yet
var InitRepository bool

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a remote backup of specified path",
	Long: `
Creates a backup of the specified path and its subdirectories and stores it in
a repository. The repository can be stored locally, or on a remote server.
Backup connects to a previously initialized repository only, unless the flag
--init is added.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if BackupPath == "" {
			return errors.New("No backup path provided")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error { return Backup(BackupPath, InitRepository, Host) }
		lib.HandleCmd(f, "Error running backup", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

func addBackupOptions(c *cobra.Command) {
	f := c.Flags()
	f.BoolVar(&InitRepository, "init", false, "initialize the repository if it does not exist yet")
	f.StringVarP(&BackupPath, "path", "p", "", "local path to backup")
	viper.BindPFlag("backup_path", f.Lookup("path")) // bind backup path to environment variables
	BackupPath = viper.GetString("backup_path")
	f.StringVarP(&Host, "host", "H", "", "hostname to use in backups (defaults to $HOSTNAME)")
	viper.BindPFlag("host", f.Lookup("host")) // bind backup path to environment variables
	Host = viper.GetString("host")
}

// init registers the backupCmd with the rootCmd, which is managed by Cobra.
func init() {
	addBackupOptions(backupCmd)
	rootCmd.AddCommand(backupCmd)
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Backup performs a backup of the provided backup path and stores it in a restic repository. It uses the environment
// settings defined in lib.GetSupportedSecrets and lib.GetSupportedVariables.
func Backup(path string, init bool, host string) error {
	lib.Logger.Info().Msgf("Starting backup operation of path '%s'", path)

	// check if the repository is already initialized and do so if instructed
	if err := lib.ExecuteResticCmd(false, "snapshots"); err != nil {
		if init {
			lib.Logger.Info().Msg("Initializing repository for first use")
			if err := lib.ExecuteResticCmd(true, "init"); err != nil {
				return &lib.ResticError{Err: "Could not init repository", Fatal: true}
			}
		} else {
			return &lib.ResticError{Err: "Could not open repository", Fatal: true}
		}
	}

	// ensure the repository is unlocked
	if err := lib.ExecuteResticCmd(false, "unlock"); err != nil {
		return &lib.ResticError{Err: "Could not unlock repository", Fatal: true}
	}

	// execute the backup command
	args := []string{path}
	if host != "" {
		args = append(args, "--host="+host)
	}
	if err := lib.ExecuteResticCmd(true, "backup", args...); err != nil {
		return err
	}

	lib.Logger.Info().Msgf("Finished backup operation of path '%s'", path)
	return nil
}
