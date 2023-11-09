package channellogs

import (
	"encoding/json"
	"log/slog"
	"time"
	"fmt"

	"slackLogs/internal/args"
	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	"slackLogs/internal/model"
	"slackLogs/internal/constants"
)

var (
	totalLogsSize = 0
	logtype       = "ChannelDetail"
	logCount      = 0 // This variable helps to track number of logs exported in each request
)

var logs = []logclient.Logs{}
var slackToken string

type ChannelLogsHandler struct {
	Client *logclient.LogClient
}

func NewChannelLogsHandler(client *logclient.LogClient) *ChannelLogsHandler {
	return &ChannelLogsHandler{Client: client}
}

// ConversationsListResponse contains slack API successful response
// https://api.slack.com/methods/channels.list#examples
type channelsListResponse struct {
	Ok               bool            `json:"ok"`
	Channels         []model.Channel `json:"channels"`
	ResponseMetaData struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	ReqError string `json:"error"`
}

func getSlackChannelLogs(c *common.SlackClient, teamId string) (channelsListResponse, error) {
	slackClient := common.NewSlackClient(c.SlackAPIURL, c.SlackToken, c.Cursor)
	params := map[string]string{
                "team_id": teamId,
        }
	var responseData channelsListResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData, params)
	if errSlack != nil {
		return responseData, errSlack
	}
	if !responseData.Ok {
                return responseData, fmt.Errorf("Slack API error %v", responseData.ReqError)
        }
	return responseData, nil
}

func transformChannelLogs(channelLogs []model.Channel, teamName string) error {
	ts := time.Now().Unix()
	for _, l := range channelLogs {
		l.TeamName = teamName
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

func (cl *ChannelLogsHandler) ResetLogs() {
	if len(logs) > 0 {
		cl.Client.Flush(logtype, logs)
		slog.Info("Count of channelDetail logs pushed to NR", "logCount", logCount)
	}
	logs = []logclient.Logs{}
	totalLogsSize = 0
	logCount = 0
}

func (cl *ChannelLogsHandler) Collect(token string, teamId string, teamName string) error {
	slog.Info("Collecting channel deatils")
	flushInterval := args.GetInterval()
	nextCursor := ""
	logCount = 0
	slackToken = token
	for {
		c := common.NewSlackClient(constants.SlackChannelAPIURL, slackToken, nextCursor)
		// Get Channel logs
		response, err := getSlackChannelLogs(c, teamId)
		if err != nil {
			return err
		}
		// Filter required fields and add timestamp to each log
		err = transformChannelLogs(response.Channels, teamName)
		if err != nil {
			return err
		}
		// Check total collected logs size and maximum allowed logs size in a single request
		if totalLogsSize >= constants.MaxAllowed {
			cl.ResetLogs()
		}
		next := response.ResponseMetaData.NextCursor
		if next == "" {
			slog.Info("There is no next page, collected channels list. Sending logs to NR")
        		cl.ResetLogs()
			break
		}
		nextCursor = next
	}
	// Flush rest of the logs
        cl.ResetLogs()
	slog.Info("Collecting channel details : exit, next iteration starts", "flushInterval(in hours)", flushInterval)
        time.Sleep(time.Duration(flushInterval) * time.Minute)
	return nil
}

func GetChannels(token string, teamId string) ([]string, error) {
        slog.Info("Collecting channel ids")
        nextCursor := ""
	var channelList []string
	for {
		c := common.NewSlackClient(constants.SlackChannelAPIURL, token, nextCursor)
		// Get Channel logs
		response, err := getSlackChannelLogs(c, teamId)
		if err != nil {
			return channelList, err
		}
		for _, l := range response.Channels {
			channelList = append(channelList, l.ID)
		}
		next := response.ResponseMetaData.NextCursor
		if next == "" {
			slog.Debug("Done with fetching all the channels, now iterate through the channels to get conversations")
			break
		}
		nextCursor = next
	}
	return channelList, nil
}
