package constants

const (
        MaxAllowed    = 10000  // 1MB
	SlackUserAPIURL = "https://slack.com/api/users.list"
        SlackTeamInfoAPIURL = "https://slack.com/api/team.info"
        SlackBillingInfoAPIURL = "https://slack.com/api/team.billableInfo"
        SlackaccessAPIURL = "https://slack.com/api/team.accessLogs"
	SlackChannelAPIURL  = "https://slack.com/api/conversations.list"
	SlackChannelHistoryAPIURL  = "https://slack.com/api/conversations.history"
	SlackChannelRepliesAPIURL  = "https://slack.com/api/conversations.replies"
	SlackTeamsListAPIURL  = "https://slack.com/api/auth.teams.list"
	SlackAuditLogsAPIURL  = "https://api.slack.com/audit/v1/logs"
	UserEntity = "user"
	ChannelEntity = "channel"
	FileEntity  = "file"
	AppEntity = "app"
	WorkspaceEntity = "workspace"
	OtherEntity = "other"
        UserAuditLogType       = "UserAuditLog"
        ChannelAuditLogType     = "ChannelAuditLog"
        WorkspaceAuditLogType   =  "WorkspaceAuditLog"
        FileAuditLogType   = "FileAuditLog"
        AppAuditLogType   = "AppAuditLog"
        OtherAuditLogsType   = "OtherAuditLogs"
)
