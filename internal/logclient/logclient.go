package logclient

import (
	"net/http"
	"time"
	"log/slog"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"sync"
	"context"
	"io/ioutil"
	"errors"

	"slackLogs/internal/args"
)

type NRResponce struct {
        Success   bool   `json:"success"`
        Uuid      string `json:"uuid"`
        RequestId string `json:"requestId"`
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

type LogSet struct {
	Common *CommonAttr      `json:"common"`
	Logs   []Logs   	`json:"logs"`
}

type CommonAttr struct {
	Attributes map[string]string  `json:"attributes"`
}

type Logs struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type LogClient struct {
	//logMessage *LogSet
	mux        sync.Mutex
	msgSize    int
}

func NewLogClient() *LogClient {
	return &LogClient{msgSize: 0}
}

func (c *LogClient) Flush(logtype string, logs []Logs) error {
	slog.Debug("Flush: enter", "logtype", logtype)
	// Ensure we have something to do
	if len(logs) <= 0 {
		slog.Info("There are no logs to send for the", "logtype", logtype)
		// No data , no error
		return nil
	}
	logCount := len(logs)

	ls := LogSet{
		Common: &CommonAttr{
			Attributes: map[string]string{
				"source": "Slack",
				"logtype": logtype,
			},
		},
		Logs: make([]Logs, len(logs)),
	}
	c.mux.Lock()
        ls.Logs = logs
	err := c.ExportLogsToEndpoint(&ls)
	c.mux.Unlock()
	
	if err != nil {
		return err
	}
	slog.Info("Logs pushed to NR", "type", logtype, "count", logCount)
	slog.Debug("Flush: exit", "logtype", logtype)
	
	return nil
}

func handleErrorResponse(statusCode int) {
	switch statusCode {
	case http.StatusRequestEntityTooLarge:
		slog.Debug("There was an error when communicating to New Relic One. The message was too big. %v", statusCode)
	default :
		slog.Debug("There was an error when communicating to New Relic One. %v", statusCode)
	}
}


func (c *LogClient) ExportLogsToEndpoint(msg *LogSet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Marshal the body
	body, err := json.Marshal([]LogSet{*msg})
	if err != nil {
		slog.Error("Error marshaling json: %s", "error", err)
	}
	slog.Debug("Marshaled", "body", string(body))
	

	// Compress log data
	var compressedLogData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedLogData)
	_, errCompression := gzipWriter.Write(body)
	if errCompression != nil {
		slog.Debug("Error compressing log data:", errCompression)
		return errCompression
	}
	gzipWriter.Close()

	req, errRequest := http.NewRequestWithContext(ctx, "POST", args.GetNRLogEndpoint(), &compressedLogData)
	if errRequest != nil {
		slog.Debug("There was an error when communicating to New Relic One: %v.", errRequest)
		return errRequest
	}

	req.Header.Set("X-License-Key", args.GetNRApiKey())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "GZIP") 

	resp, err := HttpClient.Do(req)

	if err != nil {
		slog.Info("There was an error when creating a new client in New Relic One: %v.", err)
		return err
	} else {
		defer resp.Body.Close()
		body, errResponse := ioutil.ReadAll(resp.Body)
		if errResponse != nil {
			slog.Debug("There was an error when communicating to New Relic One: %v.", errResponse)
			return errResponse
		} else {
			if resp.StatusCode >= 300 {
				handleErrorResponse(resp.StatusCode)
				return errors.New("HTTP error")
			} else {
				var nr NRResponce
				errJson := json.Unmarshal(body, &nr)
				if errJson != nil {
					slog.Info("There was an error when parsing the response from New Relic One: %v.", errJson)
					return errJson
				}
				slog.Debug("Successfully pushed logs to NR", "Req Id",  nr.RequestId)
			}
		}
	}
	return nil
}
