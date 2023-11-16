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
			log.Fatalln("Not able to fetch teams list with the provided token, err" , err)
		}
		slog.Info("List of workspaces your org-wide app has been access ", "teamsList", teamsList)
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
