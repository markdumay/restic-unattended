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

func compareLists(t *testing.T, test string, list1 map[string]string, list2 [][]string) {
	if len(list1) != len(list2) {
		t.Errorf("%s returned incorrect number of secrets, got: %d, want: %d.", test, len(list1), len(list2))
	} else {
		// sort the second list to prepare for binary search
		results := make([]string, 0, len(list2))
		for k := range list2 {
			results = append(results, list2[k][0])
		}
		sort.Strings(results)

		// perform a binary search for each expected secret
		for k := range list1 {
			if sort.SearchStrings(results, k) == len(list2) {
				t.Errorf("%s has a missing secret: %s", test, k)
			}
		}
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
	vars := GetSupportedVariables()
	if err := mergo.Merge(&vars, secrets); err != nil {
		t.Errorf("ListVariables (all) could not retrieve variables, error: %s.", err.Error())
		return
	}
	m := NewSecretsManagerWithEnv(getMockEnvMap, t.TempDir())

	// test listing of set variables
	overview, err := m.ListVariables(false)
	if err != nil {
		t.Errorf("ListVariables (set) returned an error: %s.", err.Error())
	} else {
		compareLists(t, "ListVariables (set)", secrets, overview)
	}

	// test listing of all variables
	overview, err = m.ListVariables(true)
	if err != nil {
		t.Errorf("ListVariables (all) returned an errors: %s.", err.Error())
	} else {
		compareLists(t, "ListVariables (all)", vars, overview)
	}
}
