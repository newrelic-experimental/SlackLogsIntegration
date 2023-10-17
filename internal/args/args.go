package args

import (
	"flag"
	"log/slog"
	"log"
	"os"
	"strings"
)

var (
	nrAccount       string
	nrUrlLog        string
	fetchAccessLogs bool
	fetchChannelDetails  bool
	fetchUserLogs   bool
	fetchConversationLogs   bool
	logLevel   string
	flushInterval   int
)

func init() {
	flag.StringVar(&nrUrlLog, "logApiEndpoint", "https://log-api.newrelic.com/log/v1", "New Relic log endpoint")
	flag.BoolVar(&fetchChannelDetails, "channelDetails", false, "Fetch channel details")
	flag.BoolVar(&fetchUserLogs, "userLogs", false, "Fetch user logs")
	flag.BoolVar(&fetchAccessLogs, "accessLogs", false, "Fetch access logs")
	flag.BoolVar(&fetchConversationLogs, "conversationLogs", false, "Fetch conversation logs")
	flag.StringVar(&logLevel, "logLevel", "info", "Golang slog log level: debug | info | warn | error")
	flag.IntVar(&flushInterval, "flushInterval", 1440, "Flush interval in minutes")


	flag.Parse()
	if v, ok := os.LookupEnv("INGEST_KEY"); ok {
		slog.Debug("IngestKey found in env", "key", v)
		nrAccount = v
	}

	if nrAccount == "" {
		log.Fatalln("****  Please set INGEST_KEY. *****")
	}

	if !fetchChannelDetails && !fetchUserLogs && !fetchAccessLogs && !fetchConversationLogs {
		log.Fatalln("Not received log types to fetch logs. Nothing to do.")
	}

	// Setup slog
	var programLevel = new(slog.LevelVar) // Info by default
   	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
   	slog.SetDefault(slog.New(h))
   	switch strings.ToLower(logLevel) {
   	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "error":
		programLevel.Set(slog.LevelError)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	default:
		programLevel.Set(slog.LevelInfo)
   	}

}

func GetNRApiKey() string {
	return nrAccount
}

func GetNRLogEndpoint() string {
	return nrUrlLog
}

func GetInterval() int {
	return flushInterval
}

func GetAccessLogsEnabled() bool {
	return fetchAccessLogs
}

func GetChannelDetailsEnabled() bool {
	return fetchChannelDetails
}

func GetConversationLogsnabled() bool {
	return fetchConversationLogs
}

func GetUserLogsEnabled() bool {
	return fetchUserLogs
}

func GetLogLevel() string {
	return logLevel
}
