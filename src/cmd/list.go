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

// ListAll instructs the list command to display all available variables, instead of only the set variables (default).
var ListAll bool

// listCmd represents the list command. It prints an overview of supported environment variables.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the supported environment variables",
	Long: `
restic-unattended supports several environment variables on top of the default
variables supported by restic. The additional variables typically end with a
"_FILE" suffix. When initialized, restic-unattended reads the value from the 
specified variable file and maps it to the associated variable. This allows the
initialization of Docker secrets as regular environment variables, restricted
to the current process environment. Typically Docker secrets are mounted to the 
/run/secrets path, but this is not a prerequisite.
`,
	Run: func(cmd *cobra.Command, args []string) {
		f := func() error {
			m := lib.NewSecretsManager()
			return m.List(ListAll)
		}
		lib.HandleCmd(f, "Error listing variables", false)
	},
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// init registers the listCmd with the rootCmd, which is managed by Cobra. It defines an optional flag called '--all',
// which instructs the list command to display all available variables, instead of only the set variables.
func init() {
	listCmd.Flags().BoolVarP(&ListAll, "all", "a", false, "Display all available variables")
	rootCmd.AddCommand(listCmd)
}
