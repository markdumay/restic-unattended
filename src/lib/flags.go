// Copyright © 2022 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"regexp"

	"github.com/spf13/pflag"
)

type filterFlag func(flag *pflag.Flag)

//======================================================================================================================
// Public Functions
//======================================================================================================================

// GetCLIFlag returns a Flag object as flattened command-line argument.
func GetCLIFlag(flags *pflag.FlagSet, flag *pflag.Flag) ([]string, error) {
	var args []string

	switch flag.Value.Type() {
	case "stringArray":
		arr, err := flags.GetStringArray(flag.Name)
		if err != nil {
			return nil, err
		}
		for _, v := range arr {
			args = append(args, fmt.Sprintf("--%s=%s", flag.Name, v))
		}

	default:
		args = []string{fmt.Sprintf("--%s=%s", flag.Name, flag.Value.String())}
	}

	return args, nil
}

// ParseArgs converts flags to flattened command-line arguments. If the match is specified, only matching flag names are
// converted.
func ParseArgs(flags *pflag.FlagSet, match string) ([]string, error) {
	var args = []string{}
	var eval filterFlag
	var parseErr error

	if match != "" {
		// compile a regular expression to match flags by name
		re, err := regexp.Compile(match)
		if err != nil {
			return nil, &ResticError{Err: "Could not parse regular expression to match flags", Fatal: true}
		}

		// setup a function to convert a flag to an argument matching the regular expression
		eval = func(flag *pflag.Flag) {
			// stop processing additional flags if there was an error
			if parseErr != nil {
				return
			}
			// process flags
			if re.MatchString(flag.Name) {
				v, err := GetCLIFlag(flags, flag)
				if err != nil {
					parseErr = err
					return
				}
				args = append(args, v...)
			}
		}
	} else {
		// setup a a simple function to convert a flag to an argument
		eval = func(flag *pflag.Flag) {
			// stop processing additional flags if there was an error
			if parseErr != nil {
				return
			}
			v, err := GetCLIFlag(flags, flag)
			if err != nil {
				parseErr = err
				return
			}
			args = append(args, v...)
		}

	}

	// convert the flags to arguments
	flags.Visit(eval)
	if parseErr != nil {
		return nil, &ResticError{Err: "Could not parse forget arguments", Fatal: true}
	}

	return args, nil
}
