package mini

import (
    "fmt"
    "strings"

    "github.com/ronaksoft/river-msg/go/msg"
    networkCtrl "github.com/ronaksoft/river-sdk/internal/ctrl_network"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/logs"
    "github.com/ronaksoft/river-sdk/internal/minirepo"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/tools"
    "go.uber.org/zap"
)

/*
   Creation Time: 2021 - Apr - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var (
    logger *logs.Logger
)

func init() {
    logger = logs.With("MiniRiver")
}

type RiverConfig struct {
    SeedHostPorts string
    // DbPath is the path of the folder holding the sqlite database.
    DbPath string
    // DbID is used to save data for different accounts in separate databases. Could be used for multi-account cases.
    DbID string
    // MainDelegate holds all the general callback functions that let the user of this SDK
    // get notified of the events.
    MainDelegate MainDelegate

    // LogLevel
    LogLevel  int
    SentryDSN string
    // Folder path to save files
    DocumentPhotoDirectory string
    DocumentVideoDirectory string
    DocumentFileDirectory  string
    DocumentAudioDirectory string
    DocumentCacheDirectory string
    LogDirectory           string
    // ConnInfo stores the Connection Info
    ConnInfo *RiverConnection
    // ClientInfo
    ClientPlatform string
    ClientVersion  string
    ClientOs       string
    ClientVendor   string
    CountryCode    string

    MaxInFlightDownloads int32
    MaxInFlightUploads   int32

    // Team related parameters
    TeamID         int64
    TeamAccessHash int64
}

// River is the main and a wrapper around all the components of the system (networkController, queueController,
// syncController). All the controllers could be used standalone, but this SDK connect them in a way
// we think is the best possible.
// Only the functions which are exposed will be used by the user of the SDK. All the low-level tasks
// to smooth the connection between client and server are done by this SDK. The underlying storage used
// by this SDK is Badger V2. 'repo' is the package name selected to handle repository functions.
type River struct {
    ConnInfo  *RiverConnection
    dbPath    string
    sentryDSN string

    // localCommands can be satisfied by client cache
    localCommands map[int64]request.LocalHandler
    messageChan   chan []*rony.MessageEnvelope
    updateChan    chan *msg.UpdateContainer

    // Internal Controllers
    network *networkCtrl.Controller

    // Delegates
    mainDelegate MainDelegate
}

// SetConfig must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
    domain.ClientPlatform = conf.ClientPlatform
    domain.ClientVersion = conf.ClientVersion
    domain.ClientOS = conf.ClientOs
    domain.ClientVendor = conf.ClientVendor

    r.messageChan = make(chan []*rony.MessageEnvelope, 100)
    r.updateChan = make(chan *msg.UpdateContainer, 100)
    r.sentryDSN = conf.SentryDSN
    r.ConnInfo = conf.ConnInfo

    if conf.MaxInFlightDownloads <= 0 {
        conf.MaxInFlightDownloads = 10
    }
    if conf.MaxInFlightUploads <= 0 {
        conf.MaxInFlightUploads = 10
    }

    // Initialize DB Path
    conf.DbPath = strings.TrimPrefix(conf.DbPath, "file://")
    conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")
    r.dbPath = fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID)

    r.registerCommandHandlers()
    r.mainDelegate = conf.MainDelegate

    // set log level
    logger.SetLogLevel(conf.LogLevel)

    // Initialize Network Controller
    r.network = networkCtrl.New(
        networkCtrl.Config{
            SeedHosts:   strings.Split(conf.SeedHostPorts, ","),
            CountryCode: conf.CountryCode,
        },
    )
    r.network.UpdateEndpoint("")
    r.network.OnNetworkStatusChange = func(newQuality domain.NetworkStatus) {}
    r.network.OnGeneralError = r.onGeneralError
    r.network.UpdateChan = r.updateChan
    r.network.MessageChan = r.messageChan
    r.network.OnWebsocketConnect = r.onNetworkConnect

    // Initialize FileController
    repo.SetRootFolders(
        conf.DocumentAudioDirectory,
        conf.DocumentFileDirectory,
        conf.DocumentPhotoDirectory,
        conf.DocumentVideoDirectory,
        conf.DocumentCacheDirectory,
    )

    // Initialize River Connection
    logger.Info("SetConfig done!")

    // Set current team
    domain.SetCurrentTeam(conf.TeamID, uint64(conf.TeamAccessHash))
}

func (r *River) onNetworkConnect() (err error) {
    return nil
}
func (r *River) onGeneralError(requestID uint64, e *rony.Error) {
    logger.Info("received error (General)",
        zap.Uint64("ReqID", requestID),
        zap.String("Code", e.Code),
        zap.String("Item", e.Items),
    )

    if r.mainDelegate != nil && requestID == 0 {
        buff, _ := e.Marshal()
        r.mainDelegate.OnGeneralError(buff)
    }
}
func (r *River) messageReceiver() {
    // NOP Loop to just clear the received messages
    for range r.messageChan {
    }
}
func (r *River) updateReceiver() {
    // NOP Loop to just clear the received updates
    for range r.updateChan {
    }
}

func (r *River) registerCommandHandlers() {
    r.localCommands = map[int64]request.LocalHandler{
        msg.C_AccountGetTeams:        r.accountGetTeams,
        msg.C_ClientSendMessageMedia: r.clientSendMessageMedia,
        msg.C_ClientGlobalSearch:     r.clientGlobalSearch,
        msg.C_MessagesSendMedia:      r.messagesSendMedia,
        msg.C_MessagesGetDialogs:     r.messagesGetDialogs,
        msg.C_ContactsGet:            r.contactsGet,
    }
}

func (r *River) setLastUpdateID(teamID, updateID int64) error {
    return minirepo.General.SaveInt64(tools.S2B(fmt.Sprintf("%s.%d", domain.SkUpdateID, teamID)), updateID)
}

func (r *River) getLastUpdateID(teamID int64) int64 {
    return minirepo.General.GetInt64(tools.S2B(fmt.Sprintf("%s.%d", domain.SkUpdateID, teamID)))
}

func (r *River) setContactsHash(teamID int64, h uint32) error {
    return minirepo.General.SaveUInt32(tools.S2B(fmt.Sprintf("%s.%d", domain.SkContactsGetHash, teamID)), h)
}

func (r *River) getContactsHash(teamID int64) uint32 {
    return minirepo.General.GetUInt32(tools.S2B(fmt.Sprintf("%s.%d", domain.SkContactsGetHash, teamID)))
}

// RiverConnection connection info
type RiverConnection struct {
    AuthID  int64
    AuthKey []byte
    UserID  int64
}
