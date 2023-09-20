package main

import (
	"slackLogs/internal/args"
	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	"slackLogs/internal/userlogs"
	"slackLogs/internal/channellogs"
	"slackLogs/internal/accesslogs"

	"os"
	"log/slog"
	"log"
)

var logClient *logclient.LogClient
var slackToken string

func updateSlackToken() {
        val, ok := os.LookupEnv("SLACK_ACCESS_TOKEN")
        if !ok {
                log.Fatalln("**** Slack access token not set ***** ")
        }
	slackToken = val
}

func collectAndExportLogsToNR(c common.CollectLogs) {
	for  {
		err := c.Collect(slackToken)
		if err != nil {
			// Log the error
			log.Fatalln("Received an error in user logs collecting/exporting", err)
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
	// TODO: signal handling
	select {}
}
