package domain

import (
    "fmt"
    "time"

    "github.com/ronaksoft/river-msg/go/msg"
)

//go:generate go run update_version.go

var (
    ClientPlatform string
    ClientVersion  string
    ClientOS       string
    ClientVendor   string
    SysConfig      *msg.SystemConfig

    /*
    	Parameters which prevent sending duplicate requests
    */
    ContactsSynced int32
    TimeSynced     int32
    ConfigSynced   int32
)

func init() {
    // Set Default SysConfig
    SysConfig = &msg.SystemConfig{
        GifBot:                  "gif",
        WikiBot:                 "wiki",
        TestMode:                false,
        PhoneCallEnabled:        false,
        ExpireOn:                0,
        GroupMaxSize:            250,
        ForwardedMaxCount:       50,
        OnlineUpdatePeriodInSec: 90,
        EditTimeLimitInSec:      86400,
        RevokeTimeLimitInSec:    86400,
        PinnedDialogsMaxCount:   7,
        UrlPrefix:               0,
        MessageMaxLength:        4096,
        CaptionMaxLength:        4096,
        DCs:                     nil,
        MaxLabels:               20,
        TopPeerDecayRate:        3500000,
        TopPeerMaxStep:          365,
        MaxActiveSessions:       10,
        // Reactions: []string{"😂",
        //	"😢",
        //	"😡",
        //	"👎",
        //	"👍",
        //	"🙋‍♀",
        //	"🙋‍♂",
        //	"🛢",
        //	"❤",
        //	"🤐",
        //	"😖",
        //	"🙁",
        //	"\U0001F973",
        //	"🤩",
        //	"😋",
        //	"😏",
        //	"🙏",
        //	"💯",
        //	"👏",
        //	"🤝",
        //	"😍",
        //	"🤔",
        //	"🤒",
        //	"😱",
        //	"😜",
        //	"😴",
        //	"💩",
        //	"😭",
        //	"🤨",
        //	"🤦‍♀",
        //	"🙄",
        //	"✅",
        //	"💪",
        //	"👌",
        //	"🤞"},
    }
}

// Global Parameters
const (
    WebsocketIdleTimeout        = 5 * time.Minute
    WebsocketPingTimeout        = 2 * time.Second
    WebsocketWriteTime          = 3 * time.Second
    WebsocketRequestTimeout     = 3 * time.Second
    WebsocketRequestTimeoutLong = 8 * time.Second
    WebsocketDialTimeout        = 3 * time.Second
    WebsocketDialTimeoutLong    = 10 * time.Second
    HttpRequestTimeout          = 30 * time.Second
    HttpRequestTimeShort        = 8 * time.Second
    SnapshotSyncThreshold       = 10000
)

// System Keys
const (
    SkUpdateID           = "UPDATE_ID"
    SkDeviceToken        = "DEVICE_TOKEN_INFO"
    SkContactsImportHash = "CONTACTS_IMPORT_HASH"
    SkContactsGetHash    = "CONTACTS_GET_HASH"
    SkSystemSalts        = "SERVER_SALTS"
    SkReIndexTime        = "RE_INDEX_TS"
    SkGifHash            = "GIF_HASH"
    SkTeam               = "TEAM"
)

func GetContactsGetHashKey(teamID int64) string {
    return fmt.Sprintf("%s.%021d", SkContactsGetHash, teamID)
}

// NetworkStatus network controller status
type NetworkStatus int

const (
    NetworkDisconnected NetworkStatus = 0
    NetworkConnecting   NetworkStatus = 1
    NetworkConnected    NetworkStatus = 4
)

func (ns NetworkStatus) ToString() string {
    switch ns {
    case NetworkDisconnected:
        return "Disconnected"
    case NetworkConnecting:
        return "Connecting"
    case NetworkConnected:
        return "Connected"
    }
    return "Unknown"
}

// SyncStatus status of Sync Controller
type SyncStatus int

const (
    OutOfSync SyncStatus = iota
    Syncing
    Synced
)

func (ss SyncStatus) ToString() string {
    switch ss {
    case OutOfSync:
        return "OutOfSync"
    case Syncing:
        return "Syncing"
    case Synced:
        return "Synced"
    }
    return ""
}

// UserStatus Times
const (
    Minute = 60
    Hour   = Minute * 60
    Day    = Hour * 24
    Week   = Day * 7
)

// UserStatus
const (
    LastSeenUnknown = iota
    LastSeenFewSeconds
    LastSeenFewMinutes
    LastSeenToday
    LastSeenYesterday
    LastSeenThisWeek
    LastSeenLastWeek
    LastSeenThisMonth
    LastSeenLongTimeAgo
)

// Network Connection
const (
    ConnectionNone = iota
    ConnectionWifi
    ConnectionCellular
)
