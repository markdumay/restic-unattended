// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"fmt"

	"github.com/spf13/pflag"
)

//======================================================================================================================
// Public Functions
//======================================================================================================================

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
