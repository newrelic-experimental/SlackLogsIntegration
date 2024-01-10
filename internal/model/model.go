package model

// User contains all the information of a user
type User struct {
        UserID            string         `json:"id"`
        TeamID            string         `json:"team_id"`
        TeamName          string         `json:"team_name"` // This is not part of UserList repsonse.
        Billable          bool         `json:"billable"` // This is not part of UserList repsonse.
        Name              string         `json:"name"`
        Deleted           bool           `json:"deleted"`
        Color             string         `json:"color"`
        RealName          string         `json:"real_name"`
        TZ                string         `json:"tz,omitempty"`
        TZLabel           string         `json:"tz_label"`
        TZOffset          int            `json:"tz_offset"`
        IsBot             bool           `json:"is_bot"`
        IsAdmin           bool           `json:"is_admin"`
        IsOwner           bool           `json:"is_owner"`
        IsPrimaryOwner    bool           `json:"is_primary_owner"`
        IsRestricted      bool           `json:"is_restricted"`
        IsUltraRestricted bool           `json:"is_ultra_restricted"`
        IsStranger        bool           `json:"is_stranger"`
        IsAppUser         bool           `json:"is_app_user"`
        IsInvitedUser     bool           `json:"is_invited_user"`
        Has2FA            bool           `json:"has_2fa"`
        TwoFactorType     *string        `json:"two_factor_type"`
        HasFiles          bool           `json:"has_files"`
        Presence          string         `json:"presence"`
        Locale            string         `json:"locale"`
        Updated           int64       `json:"updated"`
	Random            map[string]interface{} `json:"-"`

}

// UserProfile contains all the information of a user
type userProfile struct {
        Team                   string                  `json:"team"`
        ProfileRandom          map[string]interface{}  `json:"-"`
}


// Team contains all the information of a team
type Team struct {
	Id          string                 `json:"id"`
	Name        string                 `json:"name"`
	Random      map[string]interface{} `json:"-"`
}

// Channel contains all the information of a conversation
type Channel struct {
        ID               string         `json:"id"`
        Name             string         `json:"name"`
        NumMembers       int             `json:"num_members"`
        Random           map[string]interface{} `json:"-"`
	TeamName         string            `json:"team_name"`
}

type Conversation struct {
	Type             string            `json:"type"`
	ChannelID        string
	ChannelName      string
	TeamName         string            `json:"team"`
	AppID            string            `json:"app_id"`
	Text             string            `json:"text"`
	User             string            `json:"user"`
	TimeStamp        string                     `json:"ts"`
	Blocks           []Block   `json:"blocks"`
	RepliesList        []ConversationReply
	ReplyCount       int               `json:"reply_count"`
	Random           map[string]interface{} `json:"-"`
}

type ConversationReply struct {
	Type             string         `json:"type"`
	User             string         `json:"user"`
	Text             string         `json:"text"`
	ThreadTS	 string         `json:"thread_ts"`
	ParentUserID     string         `json:"parent_user_id"`
	TimeStamp        string         `json:"ts"`
}

type Block struct {
	Type         string         `json:"type"`
	BlockID      string         `json:"block_id"`
	Elements     []map[string]interface{} `json:"elements"`
}

type ChannelSubInfo struct {
        Value    string       `json:"value"`
        Creator  string       `json:"creator"`
        LastSet  int64        `json:"last_set"`
}

// teamInfoResponse contains slack API successful response
// https://api.slack.com/methods/team.info#examples
type TeamInfoResponse struct {
        Ok               bool                   `json:"ok"`
        TeamInfo         Team                   `json:"team"`
     	ReqError         string                 `json:"error"`
	Random           map[string]interface{} `json:"-"`
}

// AccessLog contains all the access information of a user
type AccessLog struct {
        UserID     string     `json:"user_id"`
        Username   string     `json:"username"`
        DateFirst  int64      `json:"date_first"`
        DateLast   int64      `json:"date_last"`
        Count      int        `json:"count"`
        IPAddress  string     `json:"ip"`
        UserAgent  string     `json:"user_agent"`
        ISP        string     `json:"isp"`
        Country    string     `json:"country"`
        Region     string     `json:"region"`
	TeamName   string     `json:"team_name"`
        Random     map[string]interface{} `json:"-"`
}
