package conversationlogs

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
	"slackLogs/internal/channellogs"
)

var (
	totalLogsSize = 0
	logtype       = "ConversationLog"
	logCount      = 0 // This variable helps to track number of logs exported in each request
)

var logs = []logclient.Logs{}
var slackToken string
var teamName string

type ConversationLogsHandler struct {
	Client *logclient.LogClient
}

func NewConversationLogsHandler(client *logclient.LogClient) *ConversationLogsHandler {
	return &ConversationLogsHandler{Client: client}
}

// conversationsListResponse contains slack API successful response
// https://api.slack.com/methods/conversations.history#examples
type conversationsListResponse struct {
	Ok               bool         `json:"ok"`
	ConversationsList        []model.Conversation `json:"messages"`
	ResponseMetaData struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	HasMoreData bool `json:"has_more"`
	ReqError string `json:"error"`
	Random            map[string]interface{} `json:"-"`
}

// conversationsListResponse contains slack API successful response
// https://api.slack.com/methods/conversations.replies#examples
type conversationsReplyResponse struct {
	Ok               bool         `json:"ok"`
	RepliesList        []model.ConversationReply `json:"messages"`
	ResponseMetaData struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	HasMoreData bool `json:"has_more"`
	ReqError string `json:"error"`
	Random            map[string]interface{} `json:"-"`
}

// teamInfoResponse contains slack API successful response
// https://api.slack.com/methods/team.info#examples
type TeamInfoResponse struct {
	Ok               bool                   `json:"ok"`
	TeamInfo         model.Team             `json:"team"`
	Random           map[string]interface{} `json:"-"`
}

type replyParams struct {
	TimeStamp int64
	Channel   string
	Token     string  
	Latest    int64
	Oldest    int64 
	Client    *common.SlackClient
}


func getSlackConversationLogs(c *common.SlackClient, channelId string, oldest int64, latest int64) (conversationsListResponse, error) {
	params := map[string]string{
                "channel": channelId,
                "inclusive": strconv.FormatBool(true),
                "limit": strconv.Itoa(200),
                "latest": strconv.FormatInt(latest, 10),
                "oldest": strconv.FormatInt(oldest, 10),
        }
	var responseData conversationsListResponse
	errSlack := c.SendRequest(common.WaitAndRetry, &responseData, params)
	if errSlack != nil {
		return responseData, errSlack
	}
	if !responseData.Ok {
		return responseData, fmt.Errorf("Slack API error %v", responseData.ReqError)
	}
	return responseData, nil
}

func transformConversationLogs(conversationLogs []model.Conversation, channelID string) error {
	ts := time.Now().Unix()
	for _, l := range conversationLogs {
		if l.ReplyCount >= 1 {
			repliesList, err := getReplies(l.TimeStamp, channelID)	
			if err != nil {
				return fmt.Errorf("Error while getting replies for channel %s -  %v", channelID, err)
			}
			l.RepliesList = repliesList
		}
		l.TeamName = teamName
		l.ChannelID = channelID
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

func (cl *ConversationLogsHandler) ResetLogs() {
	if len(logs) > 0 {
		cl.Client.Flush(logtype, logs)
		slog.Info("Count of conversationLogs logs pushed to NR", "logCount", logCount)
	}
	logs = []logclient.Logs{}
	totalLogsSize = 0
        logCount = 0
}

func getReplies(timeStamp string, channelId string) ([]model.ConversationReply, error) {
	nextCursor := ""
	var repliesList []model.ConversationReply
	ts := timeStamp
	for {
		slackClient := common.NewSlackClient(constants.SlackChannelRepliesAPIURL, slackToken, nextCursor)
		params := map[string]string {
                          "channel": channelId,
                           "limit": strconv.Itoa(200),
                           "ts": ts,
                }
		var responseData conversationsReplyResponse 
        	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData, params)
        	if errSlack != nil {
                	return repliesList, errSlack
       		}			
        	if !responseData.Ok {
                	return repliesList, fmt.Errorf("Slack API error %v", responseData.ReqError)
        	}
		if len(repliesList) > 0 {
			 ts = repliesList[len(repliesList)-1].TimeStamp
		}
		for  _, reply := range responseData.RepliesList {
			repliesList = append(repliesList, reply)
		}
                if !responseData.HasMoreData {
                        break
                }
		next := responseData.ResponseMetaData.NextCursor
		if next == "" {
                        slog.Debug("There is no next page, in getting replies")
                        break
		}
	}
	return repliesList,nil
}


func (cl *ConversationLogsHandler) Collect(token string, teamId string, teamName string) error {
	slog.Info("Collecting conversation logs")
	flushInterval := args.GetInterval()
	nextCursor := ""
	logCount = 0
	channelList, err := channellogs.GetChannels(token, teamId)
	if err != nil {
		return err
	}
	slackToken = token
	currentTime := time.Now()
        latestTimeStamp := currentTime.Unix()
	interval := time.Duration(flushInterval)
	// If flushInterval is 24 hours , it will fetch last 24hours conversations in the channel
        oldestTimeStamp := currentTime.Add(-(interval) * time.Minute).Unix()
	for  _, channelId := range channelList {
		for {
			c := common.NewSlackClient(constants.SlackChannelHistoryAPIURL, token, nextCursor)
			// Get Conversation logs
			response, err := getSlackConversationLogs(c, channelId, oldestTimeStamp, latestTimeStamp)
			if err != nil {
				return err
			}
			// Filter required fields and add timestamp to each log
			err = transformConversationLogs(response.ConversationsList, channelId)
			if err != nil {
				return err
			}
			// Check total collected logs size and maximum allowed logs size in a single request
			if totalLogsSize >= constants.MaxAllowed {
				cl.ResetLogs()
			}
			next := response.ResponseMetaData.NextCursor
			if next == "" {
				slog.Debug("Conversation logs fetched for", "Channel", channelId)
				nextCursor = ""
				break
			}
			nextCursor = next
		}
	}
	// Flush rest of the logs
	cl.ResetLogs()
	slog.Info("Collecting conversation logs : exit,  next iteration starts", "flushInterval", flushInterval)
        time.Sleep(time.Duration(flushInterval) * time.Minute)
	return nil
}
