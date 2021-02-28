// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/markdumay/restic-unattended/lib"
)

func TestVersionInfo(t *testing.T) {
	versionInfo := VersionInfo()
	var expectation string

	versionFile := path.Join(lib.SourcePath(), "VERSION")
	if _, err := os.Stat(versionFile); err == nil {
		if version, err := lib.ReadLine(versionFile); err == nil {
			expectation = fmt.Sprintf("%s-src", version)
		}
	} else {
		t.Errorf("VersionInfo was incorrect, missing VERSION file")
	}

	if versionInfo != expectation {
		t.Errorf("VersionInfo was incorrect, got: '%s' want '%s'.", versionInfo, expectation)
	}
}
