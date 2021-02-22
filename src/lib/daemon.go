// Copyright © 2021 Mark Dumay. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

package lib

import (
	"os"
	"os/signal"
	"time"

	"github.com/robfig/cron/v3"
)

// Job defines a single cron job with a cron specification and callback function.
type Job struct {
	Tag  string
	Spec string
	RunE func() error
}

// Result represents a typed goroutine result.
type Result int

// Defines a pseudo enumeration of possible result types for a goroutine result.
const (
	Done int = iota
	Interrupted
	Error
	Fatal
)

type workerResult struct {
	result Result
	err    error
}

//======================================================================================================================
// Private Functions
//======================================================================================================================

// worker processes jobs available on the provided jobChan one at a time. The function runs indefinitely, unless
// interrupted (a signal becomes available on the sigChan). The result channel captures the reason for the worker being
// stopped, if haltOnError is set to true.
func worker(jobChan <-chan Job, sigChan <-chan os.Signal, result chan workerResult, haltOnError bool) {
	// wait for an interrupt or new available job; split into two selects to prioritize interrupts over new jobs
	for {
		select {
		case <-sigChan:
			Logger.Warn().Msg("Worker processing canceled")
			var r workerResult
			r.result = Result(Interrupted)
			result <- r
			return
		default:
		}

		select {
		case job := <-jobChan:
			Logger.Debug().Msgf("Worker '%s' started processing new job", job.Tag)
			if err := job.RunE(); err != nil {
				Logger.Error().Err(err).Msgf("Could not process worker '%s'", job.Tag)
				if haltOnError {
					var r workerResult
					r.result = Result(Error)
					r.err = err
					result <- r
					return
				}
			}
			Logger.Debug().Msgf("Worker '%s' finished processing", job.Tag)
		default:
			// suspend processing to handle any interrupts
			time.Sleep(time.Second)
		}
	}
}

// cronParser generates a parser for cron schedules. It supports optional seconds next to the commonly supported cron
// fields.
func cronParser() cron.Parser {
	return cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow |
		cron.Descriptor)
}

//======================================================================================================================
// Public Functions
//======================================================================================================================

// IsValidCron validates if a cron specification can be parsed successfully. It supports fields for minutes, hours,
// day of month, months, day of week, and optional seconds. Next to that, descriptors such as @monthly, @weekly, etc.
// are supported too. The function returns nil if the specification is valid, or a descriptive error message
// otherwise.
func IsValidCron(spec string) error {
	specParser := cronParser()
	_, err := specParser.Parse(spec)

	return err
}

// RunCron schedules one job according to a cron specification. It is a wrapper for RunCronJobs.
func RunCron(job Job, haltOnError bool) {
	RunCronJobs([]Job{job}, haltOnError)
}

// RunCronJobs schedules one or more jobs according to a cron specification. The specification supports default cron
// expressions, as well as optional seconds. See https://pkg.go.dev/gopkg.in/robfig/cron.v3 for additional
// information. The cron jobs runs indefinitely, unless interrupted (e.g. pressing Ctrl-C or sending SIGINT). Use the
// the callback function cmd of each job to execute a specific command at the defined interval.
//
// Jobs run one at a time and are delayed if the previous job is still running. As the cron package does not support
// chaining across different jobs, each cron job is processed by a single worker routine using a dedicated job channel.
// Jobs are added to this channel once they are released by the cron scheduler. The channel has a maximum capacity of 5
// jobs, additional jobs are dropped. The worker routine supports graceful termination.
func RunCronJobs(jobs []Job, haltOnError bool) error {
	// capture interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// setup cron processing, delaying execution if a previous job is still running
	// jobs are dropped when a backlog of 5 items is reached
	jobChan := make(chan Job, 5)
	cron := cron.New(cron.WithParser(cronParser()))
	for _, j := range jobs {
		// copy job value to avoid reuse of loop variables across goroutines
		// see: https://golang.org/doc/effective_go.html?h=panic#channels
		job := j

		Logger.Info().Msgf("Scheduling job '%s' with cron spec '%s'", job.Tag, job.Spec)
		wrapper := func() {
			// put job in the channel unless it is full
			select {
			case jobChan <- job:
				Logger.Debug().Msgf("Added new job '%s' to channel", job.Tag)
			default:
				Logger.Error().Msgf("Dropped job '%s' (channel is full)", job.Tag)
			}

		}
		id, err := cron.AddFunc(job.Spec, wrapper)
		if err != nil {
			Logger.Error().Msgf("Could not schedule job '%s'", job.Tag)
		} else {
			entry := cron.Entry(id)
			t := entry.Schedule.Next(time.Now()).Format(time.RFC3339)
			Logger.Info().Msgf("First '%s' job scheduled to run at '%s'", job.Tag, t)
		}
	}

	// setup a deferred clean-up function
	defer func() {
		cron.Stop()
		signal.Stop(sigChan)
		close(jobChan)
		Logger.Debug().Msg("Exiting lib.RunCronJobs()")
	}()

	// start the worker and cron scheduler
	result := make(chan workerResult)
	go worker(jobChan, sigChan, result, haltOnError)
	cron.Start()

	// wait for the worker and terminate on error
	r := <-result
	switch r.result {
	case Result(Interrupted):
		return &ResticError{"Cron processing interrupted", false}
	case Result(Error):
		return &ResticError{"Error processing cron jobs", false}
	case Result(Fatal):
		return &ResticError{"Error processing cron jobs", true}
	default:
		return nil
	}
}
