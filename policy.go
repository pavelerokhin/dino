package main

import (
	"math"
	"net/http"
	"time"
)

// Backoff specifies a policy how long to wait between retries
type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

// CheckForRetry specifies the policy for handling retries
type CheckForRetry func(resp *http.Response, err error) (bool, error)

// defaultBackoffPolicy provides a default callback for client.CheckBackoff, which will perform exponential
// backoff based on attempt number
// and limited by the provided minimum and maximum durations
func defaultBackoffPolicy(min, max time.Duration, attemptNum int, _ *http.Response) time.Duration {
	delay := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(delay)

	if float64(sleep) != delay || sleep > max {
		sleep = max
	}

	return sleep
}

// defaultRetryPolicy provides a default callback for client.checkForRetry, which will retry on connection
// and server errors
func defaultRetryPolicy(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return true, nil
	}

	return false, nil
}
