package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	mega = 1024
)

type DataClient struct {
	backoff       Backoff       // backoff specifies the policy for how long to wait between retries
	checkForRetry CheckForRetry // a policy for handling retries, and is called after each request
	httpClient    *http.Client
	retryWaitMin  time.Duration // minimum time to wait
	retryWaitMax  time.Duration // maximum time to wait
	retryMax      int           // maximum number of retries
	timeout       time.Duration // timeout for the request
}

func (c *DataClient) do(req *Request) (*http.Response, error) {
	for i := 0; i < c.retryMax; i++ {
		var code int // HTTP response code
		// always rewind the request body when non-nil
		if req.body != nil {
			if _, err := req.body.Seek(0, 0); err != nil {
				return nil, fmt.Errorf("failed to seek body: %v", err)
			}
		}
		// attempt the request
		resp, err := c.httpClient.Do(req.Request)
		if err != nil {
			return nil, err
		}

		code = resp.StatusCode

		// it should continue with retries?
		checkOk, checkErr := c.checkForRetry(resp, err)

		// decide if we should continue
		if !checkOk {
			if checkErr != nil {
				err = checkErr
			}

			return resp, err
		}

		// we're going to retry
		drainBody(resp.Body)

		remain := c.retryMax - i

		if remain == 0 {
			break
		}

		wait := c.backoff(c.retryWaitMin, c.retryWaitMax, i, resp)
		desc := fmt.Sprintf("%s %s", req.Method, req.URL)

		if code > 0 {
			desc = fmt.Sprintf("%s (status: %d) ", desc, code)
		}

		time.Sleep(wait)
	}

	// return an error if we fall filtering of the retry loop
	return nil, fmt.Errorf("%s %s giving up after %d attempts", req.Method, req.URL, c.retryMax)
}

func (c *DataClient) NewRequest(method, url string, body io.ReadSeeker) (*Request, error) {
	// wrap the body in a no-op ReadCloser if non-nil. This prevents the reader from being closed by the HTTP client
	var rcBody io.ReadCloser
	if body != nil {
		rcBody = io.NopCloser(body)
	}

	httpReq, err := http.NewRequest(method, url, rcBody)
	if err != nil {
		return nil, errors.Join(err, errors.New("cannot create request"))
	}

	return &Request{body, httpReq}, nil
}

func drainBody(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, mega))
	_ = body.Close()
}
