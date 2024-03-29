// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"bufio"
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/imdario/mergo"
	"github.com/olekukonko/tablewriter"
)

// EnvMap defines a function type to retrieve environment variables as key/value pairs in a map.
type EnvMap func(folder string) map[string]string

// SecretsManager reads and stages secrets from the current environment. It supports reading of file-based secrets too,
// such as Docker secrets mounted to the /run/secrets path. By convention, file-based Docker secrets have a '_FILE'
// suffix in their variable name.
type SecretsManager struct {
	getEnvMap EnvMap
	folder    string
}

//======================================================================================================================
// Constants
//======================================================================================================================

// GetSupportedSecrets defines the supported environment variables to be initialized as Docker secret.
func GetSupportedSecrets() map[string]string {
	return map[string]string{
		"RESTIC_REPOSITORY_FILE":                "Name of file containing the repository location",
		"RESTIC_PASSWORD_FILE":                  "Name of file containing the restic password",
		"AWS_ACCESS_KEY_ID_FILE":                "Name of file containing the Amazon S3 access key ID",
		"AWS_SECRET_ACCESS_KEY_FILE":            "Name of file containing the Amazon S3 secret access key",
		"ST_USER_FILE":                          "Name of file containing the Username for keystone v1 authentication",
		"ST_KEY_FILE":                           "Name of file containing the Password for keystone v1 authentication",
		"OS_USERNAME_FILE":                      "Name of file containing the Username for keystone authentication",
		"OS_PASSWORD_FILE":                      "Name of file containing the Password for keystone authentication",
		"OS_TENANT_ID_FILE":                     "Name of file containing the Tenant ID for keystone v2 authentication",
		"OS_TENANT_NAME_FILE":                   "Name of file containing the Tenant name for keystone v2 authentication",
		"OS_USER_DOMAIN_NAME_FILE":              "Name of file containing the User domain name for keystone authentication",
		"OS_PROJECT_NAME_FILE":                  "Name of file containing the Project name for keystone authentication",
		"OS_PROJECT_DOMAIN_NAME_FILE":           "Name of file containing the Project domain name for keystone authentication",
		"OS_APPLICATION_CREDENTIAL_ID_FILE":     "Name of file containing the Application Credential ID (keystone v3)",
		"OS_AUTH_TOKEN_FILE":                    "Name of file containing the Auth token for token authentication",
		"OS_APPLICATION_CREDENTIAL_NAME_FILE":   "Name of file containing the Application Credential Name (keystone v3)",
		"OS_APPLICATION_CREDENTIAL_SECRET_FILE": "Name of file containing the Application Credential Secret (keystone v3)",
		"B2_ACCOUNT_ID_FILE":                    "Name of file containing the Account ID or applicationKeyId for Backblaze B2",
		"B2_ACCOUNT_KEY_FILE":                   "Name of file containing the Account Key or applicationKey for Backblaze B2",
		"AZURE_ACCOUNT_NAME_FILE":               "Name of file containing the Account name for Azure",
		"AZURE_ACCOUNT_KEY_FILE":                "Name of file containing the Account key for Azure",
		"GOOGLE_PROJECT_ID_FILE":                "Name of file containing the Project ID for Google Cloud Storage",
	}
}

// GetSupportedVariables defines the environment variables supported by restic by default.
func GetSupportedVariables() map[string]string {
	return map[string]string{
		"RESTIC_LOGLEVEL":                  "Level of logging to use: panic, fatal, error, warn, info, debug, trace",
		"RESTIC_TIMESTAMP":                 "Timestamp (RFC 3339) prefix for each log message (schedule defaults to true)",
		"RESTIC_BACKUP_PATH":               "Local path to backup",
		"RESTIC_HOST":                      "Hostname to use in backups (defaults to $HOSTNAME)",
		"RESTIC_REPOSITORY":                "Location of the repository",
		"RESTIC_PASSWORD":                  "The actual password for the repository",
		"RESTIC_PASSWORD_COMMAND":          "Command printing the password for the repository to stdout",
		"RESTIC_KEY_HINT":                  "ID of key to try decrypting first, before other keys",
		"RESTIC_CACHE_DIR":                 "Location of the cache directory",
		"RESTIC_PROGRESS_FPS":              "Frames per second by which the progress bar is updated",
		"TMPDIR":                           "Location for temporary files",
		"AWS_ACCESS_KEY_ID":                "Amazon S3 access key ID",
		"AWS_SECRET_ACCESS_KEY":            "Amazon S3 secret access key",
		"AWS_DEFAULT_REGION":               "Amazon S3 default region",
		"ST_AUTH":                          "Auth URL for keystone v1 authentication",
		"ST_USER":                          "Username for keystone v1 authentication",
		"ST_KEY":                           "Password for keystone v1 authentication",
		"OS_AUTH_URL":                      "Auth URL for keystone authentication",
		"OS_REGION_NAME":                   "Region name for keystone authentication",
		"OS_USERNAME":                      "Username for keystone authentication",
		"OS_PASSWORD":                      "Password for keystone authentication",
		"OS_TENANT_ID":                     "Tenant ID for keystone v2 authentication",
		"OS_TENANT_NAME":                   "Tenant name for keystone v2 authentication",
		"OS_USER_DOMAIN_NAME":              "User domain name for keystone authentication",
		"OS_PROJECT_NAME":                  "Project name for keystone authentication",
		"OS_PROJECT_DOMAIN_NAME":           "Project domain name for keystone authentication",
		"OS_APPLICATION_CREDENTIAL_ID":     "Application Credential ID (keystone v3)",
		"OS_APPLICATION_CREDENTIAL_NAME":   "Application Credential Name (keystone v3)",
		"OS_APPLICATION_CREDENTIAL_SECRET": "Application Credential Secret (keystone v3)",
		"OS_STORAGE_URL":                   "Storage URL for token authentication",
		"OS_AUTH_TOKEN":                    "Auth token for token authentication",
		"B2_ACCOUNT_ID":                    "Account ID or applicationKeyId for Backblaze B2",
		"B2_ACCOUNT_KEY":                   "Account Key or applicationKey for Backblaze B2",
		"AZURE_ACCOUNT_NAME":               "Account name for Azure",
		"AZURE_ACCOUNT_KEY":                "Account key for Azure",
		"GOOGLE_PROJECT_ID":                "Project ID for Google Cloud Storage",
		"GOOGLE_APPLICATION_CREDENTIALS":   "Application Credentials for Google Cloud Storage",
		"RCLONE_BWLIMIT":                   "rclone bandwidth limit",
	}
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// filter returns a filtered array or map of elements that conform to the provided test function.
func filter(t interface{}, test func(string) bool) (ret interface{}) {
	switch t := t.(type) {
	case []string:
		filtered := []string{}
		for _, value := range t {
			if test(value) {
				filtered = append(filtered, value)
			}
		}
		return filtered
	case map[string]string:
		filtered := map[string]string{}
		for key, value := range t {
			if test(key) {
				filtered[key] = value
			}
		}
		return filtered
	}

	return nil
}

// getEnvMap retrieves all environment variables as key/value pairs in a map. All keys are converted to upper case.
func getEnvMap(folder string) map[string]string {
	env := map[string]string{}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		env[strings.ToUpper(pair[0])] = pair[1]
	}
	return env
}

// readSecret returns the content of a secret file indicated by path. Only the first line of the text file is
// retrieved. It returns an error if the file cannot be found.
func readSecret(path string) (string, error) {
	return ReadLine(path)
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// NewSecretsManager creates a new secrets manager reading variables from the current environment.
func NewSecretsManager() *SecretsManager {
	return &SecretsManager{getEnvMap: getEnvMap}
}

// NewSecretsManagerWithEnv creates a new secrets manager reading variables by calling the env() function. The folder
// argument is passed to the env() function for each call.
func NewSecretsManagerWithEnv(env EnvMap, folder string) *SecretsManager {
	return &SecretsManager{getEnvMap: env, folder: folder}
}

// InitSecrets returns an array of secrets in the form "key=value". Each supported secret, identified by the  suffix
// "_FILE", is read from a mounted file. This allows initialization of Docker secrets as regular environment variables,
// restricted to the current process environment. Typically Docker secrets are mounted to the /run/secrets path, but
// this is not a prerequisite. The keys of the returned secrets are converted to upper case.
//
// For example, imagine the following environment variables is set:
//
// "B2_ACCOUNT_ID_FILE=/run/secrets/B2_ACCOUNT_ID"
//
// InitSecrets reads the first line of the file /run/secrets/B2_ACCOUNT_ID and assigns it to a new variable
// B2_ACCOUNT_ID (note the "_FILE" suffix is stripped). See GetSupportedSecrets for an overview of all supported
// environment variables.
func (s *SecretsManager) InitSecrets() (vars []string, err error) {
	// filter for supported secrets
	test := func(s string) bool {
		supported := GetSupportedSecrets()
		if _, ok := supported[strings.ToUpper(s)]; ok {
			return true
		}
		return false
	}
	env := s.getEnvMap(s.folder)
	filtered, ok := filter(env, test).(map[string]string)
	if !ok {
		return []string{}, errors.New("Secrets cannot be read")
	}

	// read supported secrets from their path
	secrets := []string{}
	for key, path := range filtered {
		secret, err := readSecret(path)
		if err != nil {
			return []string{}, errors.New("Secrets cannot be read")
		}
		newKey := strings.TrimSuffix(key, "_FILE")
		secrets = append(secrets, newKey+"="+secret)
	}

	return secrets, nil
}

// List displays an overview of all supported environment variables, both file-based secrets and variables supported by
// restic by default. The list consists of the columns "Variable", "Set", and "Description" for each variable. If
// ListAll is set, all available variables are display, otherwise only the set variables are shown.
func (s *SecretsManager) List(listAll bool) error {
	// get overview of variables
	overview, err := s.ListVariables(listAll)
	if err != nil {
		return &ResticError{Err: err.Error(), Fatal: true}
	}

	// render output using logger
	if len(overview) < 1 {
		Logger.Info().Msg("No variables defined")
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
		table.SetTablePadding("  ") // pad with spaces
		table.SetNoWhiteSpace(true)
		table.AppendBulk(overview) // add Bulk Data
		table.Render()

		// log each line of the rendered table using a scanner
		scanner := bufio.NewScanner(strings.NewReader(tableString.String()))
		for scanner.Scan() {
			Logger.Info().Msg(scanner.Text())
		}
	}

	return nil
}

// ListVariables returns a multi-dimensional array of environment variables, with three columns "Variable", "Set",
// and "Description" for each row. If listAll is set to true, all available variables are returned - both the
// variables supported by restic by default, as well as the variables additionally supported by restic-unattended.
// The variables are sorted alphabetically and do not include a header.
func (s *SecretsManager) ListVariables(listAll bool) (l [][]string, e error) {
	// retrieve all supported secrets and default variables sorted alphabetically
	vars := GetSupportedSecrets()
	if err := mergo.Merge(&vars, GetSupportedVariables()); err != nil {
		return [][]string{}, errors.New("Cannot retrieve supported variables")
	}
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// retrieve all environment variables as key/value pair
	env := s.getEnvMap(s.folder)

	// export columns "Variable", "Set", "Description" for each variable if set or asked to list all
	list := [][]string{}
	for _, key := range keys {
		isSet := "No"
		key = strings.ToUpper(key)
		if _, ok := env[key]; ok {
			isSet = "Yes"
		}
		if isSet == "Yes" || listAll {
			element := []string{}
			element = append(element, key)
			element = append(element, isSet)
			element = append(element, vars[key])
			list = append(list, element)
		}
	}

	return list, nil
}

// StageEnv stages file-based Docker secrets and merges them with the environment variables of the current process
// context. The resulting variables can be assigned to the environment of a command. It returns an error if the
// required variables (or secrets) are missing, or if the secrets cannot be read. See InitSecretsFromEnv for more
// details about the processing of Docker secrets, and ValidatePrerequisites for the tested prerequisites.
func (s *SecretsManager) StageEnv() (vars []string, e error) {
	// validate required variables are set
	if err := s.ValidatePrerequisites(); err != nil {
		return []string{}, err
	}

	// initialize the Docker secrets
	secrets, err := s.InitSecrets()
	if err != nil {
		return []string{}, err
	}

	// discard all environment variables referring to a file-based secret
	test := func(s string) bool {
		supported := GetSupportedSecrets()
		if _, ok := supported[strings.ToUpper(s)]; ok {
			return false
		}
		return true
	}

	// retrieve all environment variables as key/value pair
	env := s.getEnvMap(s.folder)
	filtered, ok := filter(env, test).(map[string]string)
	if !ok {
		return []string{}, errors.New("Environment variables cannot be read")
	}

	// merge the secrets with filtered environment variables
	results := []string{}
	for k, v := range filtered {
		results = append(results, k+"="+v)
	}
	results = append(results, secrets...)

	return results, nil
}

// ValidatePrerequisites validates if both the restic repository and password are set as environment variable. It
// returns an error if either variable is missing, or no error otherwise.
func (s *SecretsManager) ValidatePrerequisites() error {
	// retrieve all environment variables as key/value pair
	env := s.getEnvMap(s.folder)

	// check setting of RESTIC_REPOSITORY and RESTIC_REPOSITORY_FILE
	repository := false
	if _, ok := env["RESTIC_REPOSITORY"]; ok {
		repository = true
	}
	if _, ok := env["RESTIC_REPOSITORY_FILE"]; ok {
		repository = true
	}
	if !repository {
		return errors.New("Either 'RESTIC_REPOSITORY' or 'RESTIC_REPOSITORY_FILE' needs to be set")
	}

	// check setting of RESTIC_PASSWORD and RESTIC_PASSWORD_FILE
	password := false
	if _, ok := env["RESTIC_PASSWORD"]; ok {
		password = true
	}
	if _, ok := env["RESTIC_PASSWORD_FILE"]; ok {
		password = true
	}
	if !password {
		return errors.New("Either 'RESTIC_PASSWORD' or 'RESTIC_PASSWORD_FILE' needs to be set")
	}

	return nil
}
