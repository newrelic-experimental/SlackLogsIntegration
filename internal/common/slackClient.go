package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type CollectLogs interface {
	Collect(token string, teamId string, teamName string) error // A common method for collecting data.
	ResetLogs()  // A common method for cleanup
}

type SlackClient struct {
	SlackAPIURL  string
	SlackToken   string
	DefaultLimit int
	Cursor       string
}

var (
	HttpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true, // Since we're compressing ourselves
			DisableKeepAlives:   false,
		},
		Timeout: 20 * time.Second,
	}
)

func NewSlackClient(url string, token string, cursor string) *SlackClient {
	return &SlackClient{DefaultLimit: 10, SlackAPIURL: url, SlackToken: token, Cursor: cursor}
}

type RetryCallback func(resp *http.Response) bool

func WaitAndRetry(resp *http.Response) bool {
	// Check the response status code
	if resp.StatusCode == http.StatusTooManyRequests {
		retry, _ := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		slog.Debug("HTTP TooManyRequests: seconds wait to", "retry", retry)
		time.Sleep(time.Duration(retry) * time.Second)
		return true // Retry is needed
	}
	return false // No retry needed
}

func (c *SlackClient) SendRequest(retryCallback RetryCallback, responseData interface{}, optionalParams ...map[string]string) error {
	params := url.Values{}
	limited := false
	if c.Cursor != "" {
		params.Add("cursor", c.Cursor)
	}
	if len(optionalParams) > 0 {
		opts := optionalParams[0]
		for param, value := range opts {
			if param == "limit" {
				limited = true
			}
			params.Add(param, value)
		}
		if !limited {
			params.Add("limit", strconv.Itoa(c.DefaultLimit))
		}
	}

	encodedParams := params.Encode()
	slackUrl := fmt.Sprintf("%s?%s", c.SlackAPIURL, encodedParams)
	slog.Debug("API request", "slackUrl", slackUrl)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", slackUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.SlackToken)
	response, errClient := HttpClient.Do(req)
	if errClient != nil {
		return errClient
	}

	if retryCallback(response) {
		slog.Debug("Retry same request")
		// Retry the same request after a delay
		if len(optionalParams) > 0 {
			return c.SendRequest(retryCallback, responseData, optionalParams[0])
		}
		return c.SendRequest(retryCallback, responseData)
	}

	if response.StatusCode >= 300 {
		return fmt.Errorf("HTTP error %v", response.StatusCode)
	}
	defer response.Body.Close()
	body, errResponse := ioutil.ReadAll(response.Body)
	slog.Debug("Collected logs in slackAPI", "body", string(body))
	if errResponse == nil {
		if err = json.Unmarshal(body, &responseData); err != nil {
			return err
		}
		return nil
	}
	return errResponse
}
