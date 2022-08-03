// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package main

import (
	"github.com/markdumay/restic-unattended/cmd"
)

//======================================================================================================================
// Variables
//======================================================================================================================

// BuildVersion returns the content of the ./VERSION file, added at compile time.
var BuildVersion string

//======================================================================================================================
// Private Functions
//======================================================================================================================

// main is the entrypoint of the app. It initializes the app version and invokes the CLI parser (Cobra).
func main() {
	// initialize the binary version and run CLI parsing (handled by Cobra)
	cmd.BuildVersion = BuildVersion
	cmd.Execute()
}
