// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/imdario/mergo"
)

//======================================================================================================================
// Private Functions
//======================================================================================================================

// getMockEnvMap creates file-based secrets and env variables for testing. The full list returned by
// GetSupportedSecrets() is used. The secrets are created as temporary files, which get destroyed once the unit test
// is finished. The function returns a map of key/value pairs of all created secrets.
func getMockEnvMap(folder string) map[string]string {
	secrets := GetSupportedSecrets()
	env := map[string]string{}
	for secret := range secrets {
		name := strings.TrimSuffix(secret, "_FILE")
		path := path.Join(folder, name)
		WriteLine(path, name)
		env[secret] = path
	}

	return env
}

func compareKeys(t *testing.T, test string, got []string, want []string) {
	// confirm got and want have the same length
	if len(got) != len(want) {
		t.Errorf("%s returned incorrect number of keys, got: %d, want: %d.", test, len(got), len(want))
		return
	}

	// validate got and want are equal
	sort.Strings(got)
	sort.Strings(want)
	if !Equal(got, want) {
		t.Errorf("%s is missing one or more keys", test)
	}
}

func compareResults(t *testing.T, test string, got [][]string, want []string) {
	result, err := GetColumn(got, 0)
	if err != nil {
		t.Errorf("%s returned an error: %s.", test, err.Error())
	} else {
		compareKeys(t, test, result, want)
	}
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

func TestInitSecrets(t *testing.T) {
	// initialize test secrets and secrets manager
	secrets := GetSupportedSecrets()
	m := NewSecretsManagerWithEnv(getMockEnvMap, t.TempDir())

	// read the test secrets from files
	vars, err := m.InitSecrets()
	if err != nil {
		t.Errorf("InitSecrets returned an error: %s.", err.Error())
		return
	}

	// validate all secrets are accounted for
	if len(vars) != len(secrets) {
		t.Errorf("InitSecrets returned incorrect number of variables, got: %d, want: %d.",
			len(vars), len(secrets))
		return
	}

	// validate the contents of each individual secret; left of '=' sign should equal right of '=' sign
	for _, v := range vars {
		pair := strings.SplitN(v, "=", 2)
		if pair[0] != pair[1] {
			t.Errorf("InitSecrets returned an incorrect secret, got: %s, want: %s.",
				pair[0], pair[1])
		}
	}
}

func TestListVariables(t *testing.T) {
	// initialize list of test secrets, supported variables, and secrets manager
	secrets := GetSupportedSecrets()
	secretKeys := GetKeys(secrets, false)
	vars := GetSupportedVariables()
	if err := mergo.Merge(&vars, secrets); err != nil {
		t.Errorf("ListVariables (all) could not retrieve variables, error: %s.", err.Error())
		return
	}
	varKeys := GetKeys(vars, false)
	m := NewSecretsManagerWithEnv(getMockEnvMap, t.TempDir())

	// test listing of set variables
	list, err := m.ListVariables(false)
	if err != nil {
		t.Errorf("ListVariables (set) returned an error: %s.", err.Error())
	} else {
		compareResults(t, "ListVariables (set)", list, secretKeys)
	}

	// test listing of all variables
	list, err = m.ListVariables(true)
	if err != nil {
		t.Errorf("ListVariables (all) returned an error: %s.", err.Error())
	} else {
		compareResults(t, "ListVariables (all)", list, varKeys)
	}
}
	} else {
		compareLists(t, "ListVariables (all)", vars, overview)
	}
}
