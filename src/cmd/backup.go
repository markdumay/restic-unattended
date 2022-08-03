// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"

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
		f := func() error {
			r, err := lib.NewResticManager()
			if err != nil {
				return err
			}
			return r.Backup(BackupPath, InitRepository, Host)
		}
		lib.HandleCmd(f, "Error running backup", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// TODO: verify if Host and BackupPath are properly initialized; consider to move to PreRunE
func addBackupOptions(c *cobra.Command) error {
	f := c.Flags()
	f.BoolVar(&InitRepository, "init", false, "initialize the repository if it does not exist yet")
	f.StringVarP(&BackupPath, "path", "p", "", "local path to backup")
	// bind backup path to environment variables
	if err := viper.BindPFlag("backup_path", f.Lookup("path")); err != nil {
		return fmt.Errorf("Could not bind backup_path flag")
	}
	BackupPath = viper.GetString("backup_path")
	f.StringVarP(&Host, "host", "H", "", "hostname to use in backups (defaults to $HOSTNAME)")
	// bind host to environment variables
	if err := viper.BindPFlag("host", f.Lookup("host")); err != nil {
		return fmt.Errorf("Could not bind host flag")
	}
	Host = viper.GetString("host")
	return nil
}

// init registers the backupCmd with the rootCmd, which is managed by Cobra.
func init() {
	if err := addBackupOptions(backupCmd); err != nil {
		lib.Logger.Fatal().Err(err).Msg("Could not initialize backup options")
	}
	rootCmd.AddCommand(backupCmd)
}
