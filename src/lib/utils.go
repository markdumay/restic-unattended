// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"bufio"
	"os"
)

//======================================================================================================================
// Public Functions
//======================================================================================================================

// ReadLine returns the first line of a text file indicated by a path. It returns an error if the file cannot be found.
func ReadLine(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()

	return scanner.Text(), nil
}
