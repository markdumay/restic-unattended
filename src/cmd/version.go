// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"

	"github.com/markdumay/restic-unattended/lib"
	"github.com/spf13/cobra"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// versionCmd represents the version command. It displays information about the version of the software on the console.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `The "version" command displays information about the version of this software.`,
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error { return Version() }
		lib.HandleCmd(f, "Error displaying version information", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init registers the versionCmd with the rootCmd, which is managed by Cobra.
func init() {
	rootCmd.AddCommand(versionCmd)
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Version displays information about the version of this software.
func Version() error {
	v := VersionInfo()
	if v != "" {
		lib.Logger.Info().Msgf("restic-unattended version %s", VersionInfo())
		return nil
	}

	return &lib.ResticError{Err: "Version undefined", Fatal: true}
}

// VersionInfo returns the user-friendly version of the binary. When running from source (e.g. go run main.go ...), the
// content of the ../VERSION file is retrieved with a '-src' suffix. Otherwise, the build version compiled into the
// binary is returned.
func VersionInfo() string {
	if BuildVersion != "" {
		return BuildVersion
	} else if _, err := os.Stat("../VERSION"); err == nil {
		if version, err := lib.ReadLine("../VERSION"); err == nil {
			return fmt.Sprintf("%s-src", version)
		}

	}
	return ""
}
