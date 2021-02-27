// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"path"
	"strings"
	"testing"
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
