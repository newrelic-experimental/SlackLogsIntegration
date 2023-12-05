package teamslist

import (
	"fmt"

	"slackLogs/internal/common"
	"slackLogs/internal/model"
	"slackLogs/internal/constants"
)

var slackToken string

// conversationsListResponse contains slack API successful response
// https://api.slack.com/methods/auth.teams.list#examples
type teamListResponse struct {
	Ok               bool         `json:"ok"`
	TeamsList        []model.Team `json:"teams"`
	ResponseMetaData struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
	ReqError string `json:"error"`
	Random            map[string]interface{} `json:"-"`
}


// teamInfoResponse contains slack API successful response
// https://api.slack.com/methods/team.info#examples
type teamInfoResponse struct {
        Ok               bool                   `json:"ok"`
        TeamInfo         model.Team             `json:"team"`
	ReqError         string `json:"error"`
        Random           map[string]interface{} `json:"-"`
}

func GetSlackTeamList(slackToken string) ([]model.Team, error) {
	slackClient := common.NewSlackClient(constants.SlackTeamsListAPIURL, slackToken, "")
	var responseData teamListResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData)
	if errSlack != nil {
		return nil, errSlack
	}
	if !responseData.Ok {
		return nil, fmt.Errorf("Slack API error %v", responseData.ReqError)
	}
	return responseData.TeamsList, nil
}

func GetSlackTeamInfo(slackToken string) (model.Team, error) {
	slackClient := common.NewSlackClient(constants.SlackTeamInfoAPIURL, slackToken, "")
	var responseData teamInfoResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData)
	emptyInfo := model.Team{}
	if errSlack != nil {
		return emptyInfo, errSlack
	}
	if !responseData.Ok {
		return emptyInfo, fmt.Errorf("Slack API error %v", responseData.ReqError)
	}
	return responseData.TeamInfo, nil
}
