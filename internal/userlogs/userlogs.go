package userlogs

import (
	"encoding/json"
	"log/slog"
	"time"
	"fmt"

	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	"slackLogs/internal/model"
	"slackLogs/internal/constants"
)

var (
	totalLogsSize = 0
	logtype       = "UserLog"
	logCount      = 0 // This variable helps to track number of logs exported in each request
)

var logs = []logclient.Logs{}
var slackToken string

type UserLogsHandler struct {
	Client *logclient.LogClient
}

func NewUserLogsHandler(client *logclient.LogClient) *UserLogsHandler {
	return &UserLogsHandler{Client: client}
}

// usersListResponse contains slack API successful response
// https://api.slack.com/methods/users.list#examples
type usersListResponse struct {
	Ok               bool         `json:"ok"`
	UsersList        []model.User `json:"members"`
	ResponseMetaData struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
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


type billingInfo struct {
	BillingActive bool `json:"billing_active"`
}

// billableInfoResponse contains slack API successful response
// https://api.slack.com/methods/team.billableInfo
// User billable Info
type BillableInfoResponse struct {
	Ok                bool                        `json:"ok"`
	BillableInfo      map[string]billingInfo      `json:"billable_info"`
	ReqError          string                      `json:"error"`
}


func getSlackUserLogs(c *common.SlackClient, teamId string) (usersListResponse, error) {
	slackClient := common.NewSlackClient(c.SlackAPIURL, c.SlackToken, c.Cursor)
	params := map[string]string{
                "team_id": teamId,
        }
	var responseData usersListResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData, params)
	if errSlack != nil {
		return responseData, errSlack
	}
	if !responseData.Ok {
		return responseData, fmt.Errorf("Slack API error %v", responseData.ReqError)
	}
	return responseData, nil
}

func transformUserLogs(userLogs []model.User, teamName string) error {
	ts := time.Now().Unix()
	for _, l := range userLogs {
		l.TeamName = teamName
		// TODO: Disabling this call as a fix for "not_allowed_token_type" error message
		/*status, err := getBillableInfo(l.UserID)
		if err != nil {
			l.Billable = status
		}*/
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

func (ul *UserLogsHandler) ResetLogs() {
	if len(logs) > 0 {
		ul.Client.Flush(logtype, logs)
	}
	logs = []logclient.Logs{}
	totalLogsSize = 0
        logCount = 0
}


func getBillableInfo(user string) (bool, error) {
	slackClient := common.NewSlackClient(constants.SlackBillingInfoAPIURL, slackToken, "")
	params := map[string]string{
        	"user": user,
    	}
	var responseData BillableInfoResponse
        errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData, params)
        if errSlack != nil {
                return false, errSlack
        }
	if !responseData.Ok {
                return false, fmt.Errorf("Slack API error %v", responseData.ReqError)
        }
	_, billingStatus := responseData.BillableInfo[user]
        return billingStatus, nil
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

func (ul *UserLogsHandler) Collect(token string, teamId string, teamName string) error {
	slog.Info("Collecting user logs")
	nextCursor := ""
	logCount = 0
	slackToken = token
	for {
		c := common.NewSlackClient(constants.SlackUserAPIURL, slackToken, nextCursor)
		// Get User logs
		response, err := getSlackUserLogs(c, teamId)
		if err != nil {
			return err
		}
		// Filter required fields and add timestamp to each log
		err = transformUserLogs(response.UsersList, teamName)
		if err != nil {
			return err
		}
		// Check total collected logs size and maximum allowed logs size in a single request
		if totalLogsSize >= constants.MaxAllowed {
			ul.ResetLogs()
		}
		next := response.ResponseMetaData.NextCursor
		if next == "" {
			slog.Debug("There is no next page, collected userLogs")
			break
		}
		nextCursor = next
	}
	// Flush rest of the logs
	ul.ResetLogs()
	return nil
}
