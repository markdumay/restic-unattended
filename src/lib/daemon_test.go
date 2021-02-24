// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"
)

func TestIsValidCron(t *testing.T) {
	tables := []struct {
		cron    string
		isValid bool
	}{
		{"0 0,12 * * *", true},
		{"0 0/5 * * *", true},
		{"0 0/5 * * MON,WED,FRI", true},
		{"0 9-17 * * *", true},
		{"* * * * ?", true},
		{"* * * JAN-DEC *", true},
		{"0 * * * * *", true},
		{"0 * * *", false},
		{"0 * *", false},
		{"0 *", false},
		{"0", false},
		{"@yearly", true},
		{"@annually", true},
		{"@monthly", true},
		{"@weekly", true},
		{"@daily", true},
		{"@midnight", true},
		{"@hourly", true},
	}

	for _, table := range tables {
		result := false
		if err := IsValidCron(table.cron); err == nil {
			result = true
		}
		if result != table.isValid {
			t.Errorf("IsValidCron '%s' was incorrect, got: %t, want: %t.", table.cron, result, table.isValid)
		}
	}
}
