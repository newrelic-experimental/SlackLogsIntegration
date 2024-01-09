package args

import (
	"log/slog"
	"log"
	"os"
	"strings"
	"regexp"
	"strconv"
	"fmt"
	"time"
	"io/ioutil"
	"gopkg.in/yaml.v3"
)

var (
	nrAccount       string
	nrUrlLog        string
	fetchAccessLogs bool
	fetchChannelDetails  bool
	fetchUserLogs   bool
	fetchConversationLogs   bool
	fetchAuditLogs  bool
	auditLogsPollingInterval time.Duration
	conversationLogsPollingInterval time.Duration
	channelDetailsPollingInterval time.Duration
	userLogsPollingInterval time.Duration
	accessLogsPollingInterval time.Duration
	logLevel   string
	flushLogSize   int64
)

// Config struct to match the structure of the YAML file
type Config struct {
        Global             GlobalConfig          `yaml:"global"`
        ConversationLogs   LogsAttributes        `yaml:"conversationLogs"`
        ChannelDetails     LogsAttributes        `yaml:"channelDetails"`
        UserLogs           LogsAttributes        `yaml:"userLogs"`
        AccessLogs         LogsAttributes        `yaml:"accessLogs"`
        AuditLogs          LogsAttributes        `yaml:"auditLogs"`
}

type LogsAttributes struct {
        PollingInterval    string  `yaml:"pollingInterval"`
        Enabled            bool    `yaml:"enabled"`
}

type GlobalConfig struct {
        FlushLogSize     string  `yaml:"flushLogSize"`
        LogLevel         string  `yaml:"logLevel"`
	LogApiEndpoint   string  `yaml:"logAPIEndPoint"` 
}

func parseSize(sizeStr string) (int64, error) {
        re := regexp.MustCompile(`^(\d+)\s*([BKMGbkmg])?B?$`)

        matches := re.FindStringSubmatch(strings.ToUpper(sizeStr))
        if matches == nil {
                return 0, fmt.Errorf("invalid size format: %s", sizeStr)
        }

        value, err := strconv.ParseInt(matches[1], 10, 64)
        if err != nil {
                return 0, fmt.Errorf("error parsing size value: %v", err)
        }

        unit := matches[2]
        switch unit {
        case "B", "":
                // Bytes, no conversion needed
        case "K":
                value *= 1024
        case "M":
                value *= 1024 * 1024
        case "G":
                value *= 1024 * 1024 * 1024
        default:
                return 0, fmt.Errorf("unsupported size unit: %s", unit)
        }
        return value, nil
}

func parseDuration(durationStr string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([hms]+)$`)

	matches := re.FindStringSubmatch(strings.ToLower(durationStr))
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("error parsing duration value: %v", err)
	}

	unit := matches[2]
	switch unit {
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "s":
		return time.Duration(value) * time.Second, nil
	default:
		return 0, fmt.Errorf("unsupported duration unit: %s", unit)
	}
}

func init() {
	if k, ok := os.LookupEnv("INGEST_KEY"); ok {
		slog.Debug("IngestKey found in env", "key", k)
		nrAccount = k
	}

	if nrAccount == "" {
		log.Fatalln("****  Please set INGEST_KEY. *****")
	}

	// Specify the path to YAML config file
        configFilePath := "SlackConfig.yaml"

        // Read the YAML file
        yamlFile, err := ioutil.ReadFile(configFilePath)
        if err != nil {
                log.Fatalf("Error reading YAML file: %v", err)
        }

        // Parse YAML content into a Config struct
        var config Config
        err = yaml.Unmarshal(yamlFile, &config)
        if err != nil {
                log.Fatalf("Error unmarshalling YAML content: %v", err)
        }

	flushLogSize, err = parseSize(config.Global.FlushLogSize)
        if err != nil {
                log.Fatalf("Error parsing MaxSize: %v", err)
        }

        nrUrlLog = config.Global.LogApiEndpoint
        logLevel = config.Global.LogLevel
	fetchAccessLogs = config.AccessLogs.Enabled
	if (fetchAccessLogs) {
		accessLogsPollingInterval, err = parseDuration(config.AccessLogs.PollingInterval)
		if err != nil {
			log.Fatalf("Error: %v, Please provide allowed pollingInterval for AccessLogs %v", err, config.AccessLogs.PollingInterval)
		}
	}
	fetchUserLogs = config.UserLogs.Enabled
	if (fetchUserLogs) {
		userLogsPollingInterval, err = parseDuration(config.UserLogs.PollingInterval)
		if err != nil {
			log.Fatalf("Error: %v, Please provide allowed pollingInterval for UserLogs", config.UserLogs.PollingInterval)
		}
	}
	fetchConversationLogs = config.ConversationLogs.Enabled
	if (fetchConversationLogs) {
		conversationLogsPollingInterval, err = parseDuration(config.ConversationLogs.PollingInterval)
		if err != nil {
			log.Fatalf("Error: %v, Please provide allowed pollingInterval for ConversationLogs", config.ConversationLogs.PollingInterval)
		}
	}
	fetchChannelDetails = config.ChannelDetails.Enabled
	if (fetchChannelDetails) {
		channelDetailsPollingInterval, err = parseDuration(config.ChannelDetails.PollingInterval)
		if err != nil {
			log.Fatalf("Error: %v, Please provide allowed pollingInterval for ChannelDetails", config.AccessLogs.PollingInterval)
		}
	}
	fetchAuditLogs = config.AuditLogs.Enabled
	if (fetchAuditLogs) {
		auditLogsPollingInterval, err = parseDuration(config.AuditLogs.PollingInterval)
		if err != nil {
			log.Fatalf("Error: %v, Please provide allowed pollingInterval for AuditLogs", config.AuditLogs.PollingInterval)
		}
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

func GetAccessLogsEnabled() bool {
	return fetchAccessLogs
}

func GetChannelDetailsEnabled() bool {
	return fetchChannelDetails
}

func GetConversationLogsEnabled() bool {
	return fetchConversationLogs
}

func GetUserLogsEnabled() bool {
	return fetchUserLogs
}

func GetLogLevel() string {
	return logLevel
}

func GetFlushLogSize() int64 {
	return flushLogSize
}

func GetAuditLogsEnabled() bool {
	return fetchAuditLogs
}

func GetAuditLogsPollingInterval() time.Duration {
	return auditLogsPollingInterval
}

func GetUserLogsPollingInterval() time.Duration {
	return userLogsPollingInterval
}

func GetAccessLogsPollingInterval() time.Duration {
	return accessLogsPollingInterval
}

func GetConversationLogsPollingInterval() time.Duration {
	return conversationLogsPollingInterval
}

func GetChannelDetailsPollingInterval() time.Duration {
	return channelDetailsPollingInterval
}
