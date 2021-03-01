// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

// Package cmd [...]
package cmd

import (
	"fmt"

	"github.com/markdumay/restic-unattended/lib"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// cfgFile defines the path for the Cobra config file (default is $HOME/.restic-unattended.yaml)
var cfgFile string

// BuildVersion returns the current version of the binary, added at compile time. Compile the version into the binary
// using '-ldflags'. For example, the following command builds the 'restic-unattended' binary with a version derived
// from the environment variable 'BUILD_VERSION'.
//
// go build -ldflags="-w -s -X main.BinVersion=${BUILD_VERSION}" -o /app/src/restic-unattended
var BuildVersion string

var loglevel zerolog.Level = zerolog.InfoLevel
var logformat lib.LogFormat = lib.LogFormat(lib.Default)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "restic-unattended",
	Short: "Create a backup or restore from a restic repository",
	Long: `
restic-unattended is a helper utility for restic, a fast and secure backup
program. Restic supports many backends for storing backups natively, including
AWS S3, Openstack Swift, Backblaze B2, Microsoft Azure Blob Storage, and Google
Cloud Storage.

restic-unattended simplifies the use of restic through supporting environment
variables and configuration files. It can also schedule regular backups on a 
specific interval. The tool is typically run within a Docker container, where
it also supports Docker secrets.

Some examples:
restic-unattended backup /data/to/backup
Creates a backup of the source path at the remote restic location.

restic-unattended schedule '0 0,12 * * *'
Runs a scheduled backup at midnight and noon every day.

restic-unattended version
Displays the current version of the restic-unattended binary.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initFlags(cmd.Flags())
		lib.InitLogger(logformat)
		return err
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init initializes the rootCmd for CLI processing, powered by Cobra.
func init() {
	// cobra.OnInitialize(initConfig)
	initConfig()
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.restic-unattended.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info",
		"Level of logging to use: panic, fatal, error, warn, info, debug, trace")
	rootCmd.PersistentFlags().StringP("logformat", "f", "default",
		"Log format to use: default, pretty, json")

	// bind loglevel and logformat to environment variables via viper
	if err := viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel")); err != nil {
		lib.Logger.Fatal().Err(err).Msg("Could not bind loglevel")
	}
	if err := viper.BindPFlag("logformat", rootCmd.PersistentFlags().Lookup("logformat")); err != nil {
		lib.Logger.Fatal().Err(err).Msg("Could not bind logformat")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// find home directory
		home, err := homedir.Dir()
		if err != nil {
			lib.Logger.Fatal().Err(err)
		}

		// search config in home directory with name ".restic-unattended" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".restic-unattended")
	}

	viper.SetEnvPrefix("restic")
	viper.AutomaticEnv() // read in environment variables that match

	// if a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		lib.Logger.Info().Msgf("Using config file: %s", viper.ConfigFileUsed())
	}
}

// initFlags validates the provided persistent flags and initializes applicable global values. Currently supported flags
// are "loglevel" and "logformat". They can both be provided as environment variable too.
func initFlags(flags *pflag.FlagSet) error {
	var ret error
	envLevel := viper.GetString("loglevel")
	format := viper.GetString("logformat")

	// if set, parse loglevel and initialize logger
	level, err := zerolog.ParseLevel(envLevel)
	if err != nil {
		ret = fmt.Errorf("Invalid log level '%s'", envLevel)
	} else {
		loglevel = level
		zerolog.SetGlobalLevel(loglevel)
	}

	// if set, parse logformat and initialize logger
	f, err := lib.ParseFormat(format)
	if err != nil {
		ret = fmt.Errorf("Invalid log format '%s'", format)
	} else {
		logformat = f
		lib.InitLogger(logformat)
	}

	return ret
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		lib.Logger.Fatal().Err(err)
	}
}
