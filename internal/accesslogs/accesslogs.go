package accesslogs

import (
	"encoding/json"
	"log/slog"
	"time"
	"fmt"
	"strconv"

	"slackLogs/internal/args"
	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	"slackLogs/internal/model"
	"slackLogs/internal/constants"
)

var (
	totalLogsSize = 0
	logtype       = "AccessLog"
	logCount      = 0 // This variable helps to track number of logs exported in each request
)

var logs = []logclient.Logs{}
var slackToken string
var collectedLogs bool

type accessLogsHandler struct {
	Client *logclient.LogClient
}

func NewAccessLogsHandler(client *logclient.LogClient) *accessLogsHandler {
	return &accessLogsHandler{Client: client}
}

// teamAccessLogResponse contains slack API successful response
// https://api.slack.com/methods/team.accessLogs#examples
type teamAccessLogResponse struct {
        Ok               bool            `json:"ok"`
        AccessList       []model.AccessLog `json:"logins"`
        ResponseMetaData struct {
                NextCursor string `json:"next_cursor"`
        } `json:"response_metadata"`
	ReqError string `json:"error"`
	Random            map[string]interface{} `json:"-"`
}

func getSlackaccessLogs(c *common.SlackClient, before int64) (teamAccessLogResponse, error) {
	slackClient := common.NewSlackClient(c.SlackAPIURL, c.SlackToken, c.Cursor)
	params := map[string]string{
                "before": strconv.FormatInt(before, 10),
        }
	var responseData teamAccessLogResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData, params)
	if errSlack != nil {
		return responseData, errSlack
	}
	if !responseData.Ok {
		return responseData, fmt.Errorf("Slack API error %v", responseData.ReqError)
	}
	return responseData, nil
}

func transformaccessLogs(accessLogs []model.AccessLog, teamName string, lastTimeStamp int64) error {
	ts := time.Now().Unix()
	for _, l := range accessLogs {
		if l.DateLast < lastTimeStamp {
			slog.Info("Successfully fetched accessLogs for the required interval")
			// Access logs gets retrieved latest log first.
			// Using date_last as the parameter to retrive logs in flushInterval
			collectedLogs = true
		 	continue  // Continue this loop and check other logs in this current access log list
		}
		data, errJson := json.Marshal(l)
		totalLogsSize = totalLogsSize + len(data)
		if errJson != nil {
			return errJson
		}
		lm := logclient.Logs{
			Timestamp: ts,
			Message:   string(data),
		}
		logCount = logCount + 1
		logs = append(logs, lm)
	}
	return nil
}

func (al *accessLogsHandler) ResetLogs() {
	if len(logs) > 0 {
		al.Client.Flush(logtype, logs)
		slog.Info("Count of accessLogs logs pushed to NR", "logCount", logCount)
	}
	logs = []logclient.Logs{}
	totalLogsSize = 0
        logCount = 0
}

func getTeamName() (string, error) {
	slackClient := common.NewSlackClient(constants.SlackTeamInfoAPIURL, slackToken, "")
	var responseData model.TeamInfoResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData)
	if errSlack != nil {
		return "", errSlack
	}
	if !responseData.Ok {
                return "", fmt.Errorf("Slack API error %v", responseData.ReqError)
        }
	slog.Debug("getTeamName", "teamName" , responseData.TeamInfo.Name)
	return responseData.TeamInfo.Name, nil
}

func (al *accessLogsHandler) Collect(token string) error {
	slog.Info("Collecting access logs : enter")
	flushInterval := args.GetInterval()
	nextCursor := ""
	logCount = 0
	slackToken = token
	teamName, err :=  getTeamName()
	if err != nil  {
		return err
	}
	currentTime := time.Now()
	lastFetched := currentTime.Unix()
	interval := time.Duration(flushInterval)
	lastBeforeFetched := currentTime.Add(-(interval) * time.Minute).Unix()
	for {
		c := common.NewSlackClient(constants.SlackaccessAPIURL, token, nextCursor)
		// Get access logs
		response, err := getSlackaccessLogs(c, lastFetched)
		if err != nil {
			return err
		}
		// Filter required fields and add timestamp to each log
		err = transformaccessLogs(response.AccessList, teamName, lastBeforeFetched)
		if err != nil {
			return err
		}
		if collectedLogs {
			break
		}
		// Check total collected logs size and maximum allowed logs size in a single request
		if totalLogsSize >= constants.MaxAllowed {
			al.ResetLogs()
		}
		next := response.ResponseMetaData.NextCursor
		if next == "" {
			slog.Info("There is no next page, Wait for the next polling cycle to get latest accessLogs")
			break
		}
		nextCursor = next
	}
	// Flush rest of the logs
	al.ResetLogs()
	slog.Info("Collecting access logs : exit, next iteration starts", "flushInterval", flushInterval)
        time.Sleep(time.Duration(flushInterval) * time.Minute)
	return nil
}
