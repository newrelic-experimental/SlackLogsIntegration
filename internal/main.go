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
	"slackLogs/internal/auditlogs"

	"os"
	"log/slog"
	"log"
)

var logClient *logclient.LogClient
var slackToken string
var teamsInfo = make(map[string]string)

func updateSlackToken() {
        val, ok := os.LookupEnv("SLACK_ACCESS_TOKEN")
        if !ok {
                log.Fatalln("**** Please set SLACK_ACCESS_TOKEN. ***** ")
        }
	slackToken = val
}

func collectAndExportLogsToNR(c common.CollectLogs) {
	for  {
		for id, name := range teamsInfo {
			err := c.Collect(slackToken, id, name)
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

	// Collect team information
	teamsList, err := teamslist.GetSlackTeamList(slackToken)
        if err != nil {
		log.Fatalln("Not able to fetch teams list with the provided token, err" , err)
        }
	if len(teamsList) > 0 {
		for _, team := range teamsList {
			teamsInfo[team.Id] = team.Name
		}
	} else {
		teamInfo, _ := teamslist.GetSlackTeamInfo(slackToken)
		teamsInfo[teamInfo.Id] = teamInfo.Name
	}
	slog.Info("Starting Slack API logs collection for", "teamsInfo", teamsInfo)
	if args.GetUserLogsEnabled() {
		slog.Info("UserLogs enabled: Initiating Slack API logs collection for UserLogs")
		go collectAndExportLogsToNR(userlogs.NewUserLogsHandler(logClient))
	}
	if args.GetChannelDetailsEnabled() {
		slog.Info("ChannelDetails enabled: Initiating Slack API logs collection for ChannelDetails")
		go collectAndExportLogsToNR(channellogs.NewChannelLogsHandler(logClient))
	}
	if  args.GetAccessLogsEnabled() {
		slog.Info("AccessLogs enabled: Initiating Slack API logs collection for AccessLogs")
		go collectAndExportLogsToNR(accesslogs.NewAccessLogsHandler(logClient))
	}
	if  args.GetConversationLogsnabled() {
		slog.Info("ConversationLogs enabled: Initiating Slack API logs collection for ConversationLogs")
		go collectAndExportLogsToNR(conversationlogs.NewConversationLogsHandler(logClient))
	}
	if  args.GetAuditLogsEnabled() {
		slog.Info("AuditLogs enabled: Initiating Slack API logs collection for AuditLogs")
		go collectAndExportLogsToNR(auditlogs.NewAuditLogsHandler(logClient))
	}
	// TODO: signal handling
	select {}

	slog.Info("Exiting Slack API logs collection")
}
