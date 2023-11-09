package main

import (
	"slackLogs/internal/args"
	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	"slackLogs/internal/userlogs"
	"slackLogs/internal/channellogs"
	"slackLogs/internal/accesslogs"
	"slackLogs/internal/conversationlogs"
	"slackLogs/internal/teamslist"

	"os"
	"log/slog"
	"log"
)

var logClient *logclient.LogClient
var slackToken string

func updateSlackToken() {
        val, ok := os.LookupEnv("SLACK_ACCESS_TOKEN")
        if !ok {
                log.Fatalln("**** Please set SLACK_ACCESS_TOKEN. ***** ")
        }
	slackToken = val
}

func collectAndExportLogsToNR(c common.CollectLogs) {
	for  {
		teamsList, err := teamslist.GetSlackTeamList(slackToken)
		if err != nil {
			log.Fatalln("Not able to fetch tokens list with the provided token, err" , err)
		}
		for _,team := range teamsList {
			err := c.Collect(slackToken, team.Id, team.Name)
			if err != nil {
				// Log the error
				log.Fatalln("Received an error in collecting/exporting", err)
			}
		}
	}
}

func main() {
	updateSlackToken()
	logClient = logclient.NewLogClient()

	slog.Info("Starting Slack API logs collection")
	if args.GetUserLogsEnabled() {
		go collectAndExportLogsToNR(userlogs.NewUserLogsHandler(logClient))
	}
	if args.GetChannelDetailsEnabled() {
		go collectAndExportLogsToNR(channellogs.NewChannelLogsHandler(logClient))
	}
	if  args.GetAccessLogsEnabled() {
		go collectAndExportLogsToNR(accesslogs.NewAccessLogsHandler(logClient))
	}
	if  args.GetConversationLogsnabled() {
		go collectAndExportLogsToNR(conversationlogs.NewConversationLogsHandler(logClient))
	}
	// TODO: signal handling
	select {}
}
