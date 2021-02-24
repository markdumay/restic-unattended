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

func TestRunCronJobs(t *testing.T) {
	var jobs []Job
	var logs []string
	// defined the expected results; pending on the current time, either job 1 or job 2 starts first
	result1 := []string{
		"Job 'test 1' has fired",
		"Job 'test 2' has fired",
		"Job 'test 1' has fired",
		"Job 'test 2' has fired",
	}
	result2 := []string{
		"Job 'test 2' has fired",
		"Job 'test 1' has fired",
		"Job 'test 2' has fired",
		"Job 'test 1' has fired",
	}

	// suppress all log messages unless a (fatal) error occurred
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// define the first test job, appending messages to the local log
	var test1 Job
	test1.Tag = "test 1"
	test1.Spec = "0/2 * * * * *"
	test1.Limit = 2
	test1.RunE = func() error {
		logs = append(logs, fmt.Sprintf("Job '%s' has fired", test1.Tag))
		return nil
	}
	jobs = append(jobs, test1)

	// define the second test job, appending messages to the local log
	var test2 Job
	test2.Tag = "test 2"
	test2.Spec = "1/2 * * * * *"
	test2.Limit = 2
	test2.RunE = func() error {
		logs = append(logs, fmt.Sprintf("Job '%s' has fired", test2.Tag))
		return nil
	}
	jobs = append(jobs, test2)

	// run the jobs and validate the scheduler processed the job as expected
	if err := RunCronJobs(jobs, true); err != nil {
		t.Errorf("RunCronJobs failed, error: %s", err.Error())
	}
	if !Equal(logs, result1) && !Equal(logs, result2) {
		t.Errorf("RunCronJobs failed")
	}
}
