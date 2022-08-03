// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"
)

//======================================================================================================================
// Constants and variables
//======================================================================================================================
const test1 = "This is a test string 1"
const test2 = "This is a test string 2"
const test3 = "This is a test string 3"

var inputMap = map[string]string{
	"Key 1": "Value 1",
	"Key 2": "Value 2",
	"Key 3": "Value 3",
	"Key 4": "Value 4",
	"Key 5": "Value 5",
}
var inputArray = [][]string{
	{"Key 1", "Value 1"},
	{"Key 2", "Value 2"},
	{"Key 3", "Value 3"},
	{"Key 4", "Value 4"},
	{"Key 5", "Value 5"},
}
var column1 = []string{"Key 1", "Key 2", "Key 3", "Key 4", "Key 5"}
var column2 = []string{"Value 1", "Value 2", "Value 3", "Value 4", "Value 5"}

//======================================================================================================================
// Private Functions
//======================================================================================================================

func getColumnByIndex(t *testing.T, index int, input [][]string, expected []string) error {
	result, err := GetColumn(input, index)
	if err != nil {
		return err
	}

	if len(result) != len(expected) {
		t.Errorf("GetColumn returned incorrect number of elements, got: %d, want: %d.", len(result), len(expected))
		return nil
	}

	for i, col := range result {
		if col != expected[i] {
			t.Errorf("GetColumn returned incorrect element, got: %s, want: %s.", col, expected[i])
		}
	}

	return nil
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

func TestGetKeys(t *testing.T) {
	keys := GetKeys(inputMap, true)
	if len(keys) != 5 {
		t.Errorf("GetKeys returned incorrect number of elements, got: %d, want: %d.", len(keys), 5)
	}

	for i, key := range keys {
		if key != column1[i] {
			t.Errorf("GetKeys returned incorrect key, got: %s, want: %s.", key, column1[i])
		}
	}
}

func TestGetColumn(t *testing.T) {
	if err := getColumnByIndex(t, 0, inputArray, column1); err != nil {
		t.Errorf("GetColumn returned an error: %s.", err.Error())
	}
	if err := getColumnByIndex(t, 1, inputArray, column2); err != nil {
		t.Errorf("GetColumn returned an error: %s.", err.Error())
	}
	if err := getColumnByIndex(t, 2, inputArray, column1); err == nil {
		t.Errorf("GetColumn returned unexpected result, got: nil, want: error")
	}
}

func TestEqual(t *testing.T) {
	if !Equal(column1, column1) {
		t.Errorf("Equal returned unexpected result, got: false, want: true")
	}
	if Equal(column1, column2) {
		t.Errorf("Equal returned unexpected result, got: true, want: false")
	}
}

func TestReadLine(t *testing.T) {
	path := path.Join(t.TempDir(), "test")
	if err := WriteLine(path, test1); err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
		return
	}
	result1, err := ReadLine(path)
	if err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
	}
	if result1 != test1 {
		t.Errorf("ReadLine returned unexpected result, got: %s, want: %s", result1, test1)
	}

	if err := WriteLine(path, test2); err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
	}
	if err := WriteLine(path, test3); err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
	}
	result3, err := ReadLine(path)
	if err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
	}
	if result3 != test1 {
		t.Errorf("ReadLine returned unexpected result, got: %s, want: %s", result3, test1)
	}
}

func TestWriteLine(t *testing.T) {
	path := path.Join(t.TempDir(), "test")

	if err := WriteLine(path, test1); err != nil {
		t.Errorf("WriteLine returned an error: %s", err.Error())
	}
	if err := WriteLine(path, test2); err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
	}
	if err := WriteLine(path, test3); err != nil {
		t.Errorf("ReadLine returned an error: %s", err.Error())
	}

	expected := fmt.Sprintf("%s\n%s\n%s\n", test1, test2, test3)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Errorf("WriteLine returned an error: %s", err.Error())
	} else if string(data) != expected {
		t.Errorf("WriteLine wrote incorrect file contents")
	}
}
