package auditlogs

import (
	"encoding/json"
	"log/slog"
	"time"
	"strconv"

	"slackLogs/internal/args"
	"slackLogs/internal/common"
	"slackLogs/internal/logclient"
	c"slackLogs/internal/constants"
)

var slackToken string

var (
	reset = false
)

// AuditLogEntity struct
type AuditLogEntity struct {
        Entity   string
	Logs     []logclient.Logs
        LogType   string
	LogSize  int
	LogCount int
}

// Initialize required auditlog entities
var appAuditLog = &AuditLogEntity{Entity: c.AppEntity, LogType: c.AppAuditLogType}
var channelAuditLog = &AuditLogEntity{Entity: c.ChannelEntity, LogType: c.ChannelAuditLogType}
var workspaceAuditLog = &AuditLogEntity{Entity: c.WorkspaceEntity, LogType: c.WorkspaceAuditLogType}
var fileAuditLog = &AuditLogEntity{Entity: c.FileEntity, LogType: c.FileAuditLogType}
var userAuditLog = &AuditLogEntity{Entity: c.UserEntity, LogType: c.UserAuditLogType}
var otherAuditLog = &AuditLogEntity{Entity: c.OtherEntity, LogType: c.OtherAuditLogsType}

type auditLogsHandler struct {
        Client *logclient.LogClient
}

func NewAuditLogsHandler(client *logclient.LogClient) *auditLogsHandler {
	return &auditLogsHandler{Client: client}
}

type entity struct {
	Type     string `json:"type"`
	User     user    `json:"user"` 
	Random   map[string]interface{} `json:"-"`
}

type user struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string  `json:"email"`
}

type actor struct {
	Type     string `json:"type"`
	User     user   `json:"user"`
	Random   map[string]interface{} `json:"-"`
}

type location struct {
	Type  string `json:"type"`
	Id    string  `json:"id"`
	Name  string  `json:"name"`
	Domain string  `json:"domain"` 
}

type context struct {
	Location      location `json:"location"`
	SessionId     int64   `json:"session_id"`
	Ipaddress     string  `json:"ip_address"`
	UA            string   `json:"ua"`  
	Random   map[string]interface{} `json:"-"`
}

type entry struct {
	Id               string           `json:"id"`
        DateCreate       int64            `json:"date_create"`
        Action           string           `json:"action"`
        Actor            actor            `json:"actor"`
        Entity           entity           `json:"entity"`
        Context          context          `json:"context"`
        Random           map[string]interface{} `json:"-"`
}

type AuditLogResponse struct {
	Entries          []entry          `json:"entries"`
        Random           map[string]interface{} `json:"-"`
	ResponseMetaData struct {
                NextCursor string `json:"next_cursor"`
        } `json:"response_metadata"`
}

func getSlackUserauditLogs(c *common.SlackClient, oldest int64, latest int64) (AuditLogResponse, error) {
	slackClient := common.NewSlackClient(c.SlackAPIURL, c.SlackToken, c.Cursor)
	params := map[string]string{
                "latest": strconv.FormatInt(latest, 10),
		"oldest": strconv.FormatInt(oldest, 10),
        }
	var responseData AuditLogResponse
	errSlack := slackClient.SendRequest(common.WaitAndRetry, &responseData, params)
	if errSlack != nil {
		return responseData, errSlack
	}	
	return responseData, nil
}

func transformAuditLogs(auditLogs []entry, ah *auditLogsHandler) error {
	ts := time.Now().Unix()
	for _, l := range auditLogs {
		data, errJson := json.Marshal(l)
		if errJson != nil {
			return errJson
		}
		lm := logclient.Logs{
			Timestamp: ts,
			Message:   string(data),
		}
		ah.processLogType(l.Entity.Type, lm, len(data))
	}
	return nil
}

func (ah *auditLogsHandler) processAuditLog(al *AuditLogEntity, lm logclient.Logs, size int) {
        al.LogSize = al.LogSize + size
        al.Logs = append(al.Logs, lm)
        al.LogCount = al.LogCount + 1
	ah.FlushLogs(al)
}

func (ah *auditLogsHandler) FlushLogs(al *AuditLogEntity) {
	if al.LogSize > c.MaxAllowed || reset == true {
		ah.Client.Flush(al.LogType, al.Logs)
		al.LogSize = 0
                al.Logs = nil
                al.LogCount = 0
        }
}


func (ah *auditLogsHandler) ResetLogs() {
	slog.Debug("Reset audit logs: enter")
	ah.FlushLogs(appAuditLog)
	ah.FlushLogs(userAuditLog)
	ah.FlushLogs(channelAuditLog)
	ah.FlushLogs(workspaceAuditLog)
	ah.FlushLogs(fileAuditLog)
	ah.FlushLogs(otherAuditLog)
	reset = false 
	slog.Debug("Reset audit logs: exit")
}

func (ah *auditLogsHandler) processLogType(entity string, data logclient.Logs, size int) {
	switch entity {
	case c.AppEntity:
		ah.processAuditLog(appAuditLog, data, size)
	case c.ChannelEntity:
		ah.processAuditLog(channelAuditLog, data, size)
	case c.UserEntity:
		ah.processAuditLog(userAuditLog, data, size)
	case c.FileEntity:
		ah.processAuditLog(fileAuditLog, data, size)
	case c.WorkspaceEntity:
		ah.processAuditLog(workspaceAuditLog, data, size)
	default:
		ah.processAuditLog(otherAuditLog, data, size)
	}
}

// Returns latest and oldest timestamp to collect audit logs
func getTimeRange() (int64, int64){
        currentTime := time.Now()
        lastFetched := currentTime.Unix()
	flushInterval := args.GetInterval()
        interval := time.Duration(flushInterval)
        slog.Info("Collecting audit logs", "for last(in minutes)", interval)
        lastBeforeFetched := currentTime.Add(-(interval) * time.Minute).Unix()
	return lastBeforeFetched, lastFetched
} 

func (al *auditLogsHandler) Collect(token string, teamId string, teamName string) error {
	nextCursor := ""
	slackToken = token
	oldest, latest := getTimeRange()
	
	for {
		c := common.NewSlackClient(c.SlackAuditLogsAPIURL, token, nextCursor)
		// Get audit logs
		response, err := getSlackUserauditLogs(c, oldest, latest)
		if err != nil {
			return err
		}
		// Filter audit logs based on enity type and add timestamp to each log
		err = transformAuditLogs(response.Entries, al)
		if err != nil {
			return err
		}
		next := response.ResponseMetaData.NextCursor
		if next == "" {
			slog.Debug("There is no next page, collected auditLogs")
			break
		}
		nextCursor = next
	}
	// Flush rest of the logs
	reset = true
	al.ResetLogs()
	slog.Info("Done", "Next audit logs collection iteration starts(in minutes)", args.GetInterval())
        time.Sleep(time.Duration(args.GetInterval()) * time.Minute)
	return nil
}
