// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"bufio"
	"os"
	"path"
)

//======================================================================================================================
// Public Functions
//======================================================================================================================

// Equal tells whether a and b contain the same elements. A nil argument is equivalent to an empty slice.
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

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

// WriteLine appends a line of text to a file. The file is created if it does not exist.
func WriteLine(path string, line string) error {
	// create or open the file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// append a new line and close the file
	defer file.Close()
	if _, err := file.Write([]byte(line + "\n")); err != nil {
		return err
	}

	return nil
}

// SourcePath returns the assumed main directory of the repository.
func SourcePath() string {
	if currentWorkingDirectory, err := os.Getwd(); err == nil {
		if path.Base(currentWorkingDirectory) == "cmd" {
			return path.Clean(currentWorkingDirectory + "/../..")
		} else if path.Base(currentWorkingDirectory) == "src" {
			return path.Clean(currentWorkingDirectory + "/..")
		} else {
			return currentWorkingDirectory
		}
	}
	return ""
}
