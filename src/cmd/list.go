// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"bufio"
	"strings"

	"github.com/markdumay/restic-unattended/lib"
	"github.com/olekukonko/tablewriter"
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
		f := func() error { return List() }
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

//======================================================================================================================
// Public Functions
//======================================================================================================================

// List displays an overview of all supported environment variables, both file-based secrets and variables supported by
// restic by default. The list consists of the columns "Variable", "Set", and "Description" for each variable. If
// ListAll is set, all available variables are display, otherwise only the set variables are shown.
func List() error {
	// get overview of variables
	overview, err := lib.ListVariables(ListAll)
	if err != nil {
		return &lib.ResticError{Err: err.Error(), Fatal: true}
	}

	// render output using logger
	if len(overview) < 1 {
		lib.Logger.Info().Msg("No variables defined")
	} else {
		// render a simple aligned/padded ASCII table as string
		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"Variable", "Set", "Description"})
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true)
		table.AppendBulk(overview) // add Bulk Data
		table.Render()

		// log each line of the rendered table using a scanner
		scanner := bufio.NewScanner(strings.NewReader(tableString.String()))
		for scanner.Scan() {
			lib.Logger.Info().Msg(scanner.Text())
		}
	}

	return nil
}
