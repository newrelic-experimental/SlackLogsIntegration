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

	"sync"
	"time"
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

func collectAndExportLogsToNR(c common.CollectLogs,  wg *sync.WaitGroup, logType string, iteration int) {
	defer wg.Done()
	for id, name := range teamsInfo {
		err := c.Collect(slackToken, id, name)
		if err != nil {
			// Log the error
			log.Fatalln("Received an error in collecting/exporting", err)
		}
	}
	slog.Info("Done, Collected logs", "logType", logType, "iteration", iteration)

}

func CollectLogs(interval time.Duration, c common.CollectLogs, logType string) {
	var wg sync.WaitGroup
	iteration := 1
	slog.Info("Initiating new polling iteration for", "logType", logType)
	wg.Add(1)
	go collectAndExportLogsToNR(c, &wg, logType, iteration)
	for {
		select {
		case <-time.After(interval):
			iteration++
			slog.Info("Starting polling iteration for", "logType", logType, "iteration", iteration)
			wg.Add(1)
			go collectAndExportLogsToNR(c, &wg, logType, iteration)
		}
	}
	wg.Wait()
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
		interval := args.GetUserLogsPollingInterval()
		go CollectLogs(interval, userlogs.NewUserLogsHandler(logClient), "UserLogs")
	}

	if args.GetChannelDetailsEnabled() {
		slog.Info("ChannelDetails enabled: Initiating Slack API logs collection for ChannelDetails")
		interval := args.GetChannelDetailsPollingInterval()
		go CollectLogs(interval, channellogs.NewChannelLogsHandler(logClient), "ChannelDetails")
	}

	if  args.GetAccessLogsEnabled() {
		slog.Info("AccessLogs enabled: Initiating Slack API logs collection for AccessLogs")
		interval := args.GetAccessLogsPollingInterval()
		go CollectLogs(interval, accesslogs.NewAccessLogsHandler(logClient), "AccessLogs")
	}

	if  args.GetAuditLogsEnabled() {
		slog.Info("AuditLogs enabled: Initiating Slack API logs collection for AuditLogs")
		interval := args.GetAuditLogsPollingInterval()
		go CollectLogs(interval, auditlogs.NewAuditLogsHandler(logClient), "AuditLogs")
	}

	if  args.GetConversationLogsEnabled() {
		slog.Info("ConversationLogs enabled: Initiating Slack API logs collection for ConversationLogs")
		interval := args.GetConversationLogsPollingInterval()
		go CollectLogs(interval, conversationlogs.NewConversationLogsHandler(logClient), "ConversationLogs")
	}
	select {}
	slog.Info("Exiting Slack API logs collection")
}
