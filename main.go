package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	// default retry configuration
	defaultRetryWaitMinTime = 1 * time.Second
	defaultRetryWaitMaxTime = 3 * time.Second
	defaultRetryMaxAttempts = 2
	// timeout
	defaultTimeout = 10 * time.Second
)

func New(proxyList []string) *DataClient {
	dc := DataClient{
		backoff:       defaultBackoffPolicy,
		checkForRetry: defaultRetryPolicy,
		httpClient:    http.DefaultClient,
		retryWaitMin:  defaultRetryWaitMinTime,
		retryWaitMax:  defaultRetryWaitMaxTime,
		retryMax:      defaultRetryMaxAttempts,
		timeout:       defaultTimeout,
	}
	dc.httpClient.Timeout = defaultTimeout

	if len(proxyList) > 0 {
		proxyUrl, err := url.Parse(getRandomProxyURL(proxyList))
		if err != nil {
			return &dc
		}

		dc.httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	return &dc
}

func (c *DataClient) Get(url string) (string, error) {
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read file by Url %s: %s", url, err)
	}

	return string(body), nil
}

func (c *DataClient) LocalFile(filePath string) (string, error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("cannot read file by Path %s: %s", filePath, err)
	}

	return string(body), nil
}

func (c *DataClient) Post(url string, headers map[string]string, payload []byte) (string, error) {
	requestBody := bytes.NewReader(payload)

	req, err := c.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("cannot get file by Url %s: %s", url, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read file by Url %s: %s", url, err)
	}

	return string(body), nil
}

// Request wraps the metadata needed to create HTTP requests
type Request struct {
	// body is as seekable reader over the request body payload. This is used
	// to rewind the request data in between retries
	body io.ReadSeeker
	// embed an HTTP request directly: `*Request` acts exactly like `*http.Request`
	*http.Request
}

func getRandomProxyURL(proxyList []string) string {
	return proxyList[rand.Intn(len(proxyList))]
}
