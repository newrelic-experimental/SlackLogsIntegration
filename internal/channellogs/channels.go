package channellogs

import (
	"encoding/json"
	"log/slog"
	"time"
	"fmt"

	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	"slackLogs/internal/model"
	"slackLogs/internal/constants"
	"slackLogs/internal/args"
)

var (
	totalLogsSize = 0
	logtype       = "ChannelDetail"
	logCount      = 0 // This variable helps to track number of logs exported in each request
)

var logs = []logclient.Logs{}
var slackToken string
var channelsInfo = make(map[string]string)
var isClosed = false

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
	if (!args.GetChannelDetailsEnabled()) {
		return
	}

	if len(logs) > 0 {
		cl.Client.Flush(logtype, logs)
	}
	logs = []logclient.Logs{}
	totalLogsSize = 0
	logCount = 0
}

func updateChannelsInfo(channelsListCh chan<- map[string]string) {
        channelsListCh <- channelsInfo
}

func closeChannelsInfo(channelsListCh chan<- map[string]string) {
	close(channelsListCh)
}

func (cl *ChannelLogsHandler) Collect(token string, teamId string, teamName string) error {
	slog.Info("Collecting channel deatils")
	nextCursor := ""
	logCount = 0
	slackToken = token
	slog.Info("Creating new chaneell")
	isClosed = true
	for {
		var ChannelsListCh = make(chan map[string]string)
		go RecvChannelsInfo(ChannelsListCh)
		c := common.NewSlackClient(constants.SlackChannelAPIURL, slackToken, nextCursor)
		// Get Channel logs
		response, err := getSlackChannelLogs(c, teamId)
		if err != nil {
			return err
		}
		for _, l := range response.Channels {
			channelsInfo[l.ID] = l.Name
                }
	   	ChannelsListCh <- channelsInfo
		closeChannelsInfo(ChannelsListCh)
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
			slog.Debug("There is no next page, collected channels list")
        		cl.ResetLogs()
			break
		}
		nextCursor = next
	}
	// Flush rest of the logs
        cl.ResetLogs()
	return nil
}


func RecvChannelsInfo(channelsListCh <-chan map[string]string) {
        for {
                slog.Info("Receiving the channels")
                channels, ok := <- channelsListCh
                if !ok {
                        slog.Info("Received all the channels. Closed.")
                        isClosed = true
                        return
                }
                for key, value := range channels {
                        channelsInfo[key] = value
                }
        }
}

func GetChannelsInfo() map[string]string {
        return channelsInfo
}

func GetChannelStatus() bool {
        slog.Info("GetChannelStatus", "isClosed", isClosed)
        return isClosed
}
