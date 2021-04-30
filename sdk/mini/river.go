package mini

import (
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	riversdk "git.ronaksoft.com/river/sdk/sdk/prime"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"strings"
)

/*
   Creation Time: 2021 - Apr - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func SetLogLevel(l int) {
	logs.SetLogLevel(l)
}

type RiverConfig struct {
	ServerHostPort string
	// DbPath is the path of the folder holding the sqlite database.
	DbPath string
	// DbID is used to save data for different accounts in separate databases. Could be used for multi-account cases.
	DbID string
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

// River
// This the main and a wrapper around all the components of the system (networkController, queueController,
// syncController). All the controllers could be used standalone, but this SDK connect them in a way
// we think is the best possible.
// Only the functions which are exposed will be used by the user of the SDK. All the low-level tasks
// to smooth the connection between client and server are done by this SDK. The underlying storage used
// by this SDK is Badger V2. 'repo' is the package name selected to handle repository functions.
type River struct {
	ConnInfo       *RiverConnection
	serverHostPort string
	dbPath         string
	sentryDSN      string
}

// SetConfig must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	domain.ClientPlatform = conf.ClientPlatform
	domain.ClientVersion = conf.ClientVersion
	domain.ClientOS = conf.ClientOs
	domain.ClientVendor = conf.ClientVendor

	r.sentryDSN = conf.SentryDSN
	r.ConnInfo = conf.ConnInfo
	r.serverHostPort = conf.ServerHostPort

	if conf.MaxInFlightDownloads <= 0 {
		conf.MaxInFlightDownloads = 10
	}
	if conf.MaxInFlightUploads <= 0 {
		conf.MaxInFlightUploads = 10
	}

	// Initialize DB Path
	if strings.HasPrefix(conf.DbPath, "file://") {
		conf.DbPath = conf.DbPath[7:]
	}
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")
	r.dbPath = fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID)

	// set log level
	logs.SetLogLevel(conf.LogLevel)

	// set log file path
	if conf.LogDirectory != "" {
		_ = logs.SetLogFilePath(conf.LogDirectory)
	}

	// Initialize FileController
	repo.Files.SetRootFolders(
		conf.DocumentAudioDirectory,
		conf.DocumentFileDirectory,
		conf.DocumentPhotoDirectory,
		conf.DocumentVideoDirectory,
		conf.DocumentCacheDirectory,
	)

	// Initialize River Connection
	logs.Info("River SetConfig done!")

	// Set current team
	domain.SetCurrentTeam(conf.TeamID, uint64(conf.TeamAccessHash))
}

func (r *River) ExecuteCommand(
	teamID int64, teamAccess uint64, constructor int64, commandBytes []byte, delegate riversdk.RequestDelegate,
) (requestID int64, err error) {
	if registry.ConstructorName(constructor) == "" {
		return 0, domain.ErrInvalidConstructor
	}

	// commandBytesDump := deepCopy(commandBytes)
	//
	// waitGroup := new(sync.WaitGroup)
	// requestID = domain.SequentialUniqueID()
	// logs.Debug("River executes command",
	// 	zap.String("C", registry.ConstructorName(constructor)),
	// )
	//
	// blockingMode := delegate.Flags()&RequestBlocking != 0
	// serverForce := delegate.Flags()&RequestServerForced != 0
	//
	// // if function is in blocking mode set the waitGroup to block until the job is done, otherwise
	// // save 'delegate' into delegates list to be fetched later.
	// if blockingMode {
	// 	waitGroup.Add(1)
	// 	defer waitGroup.Wait()
	// } else {
	// 	r.delegateMutex.Lock()
	// 	r.delegates[uint64(requestID)] = delegate
	// 	r.delegateMutex.Unlock()
	// }
	//
	// // Timeout Callback
	// timeoutCallback := func() {
	// 	err = domain.ErrRequestTimeout
	// 	delegate.OnTimeout(err)
	// 	releaseDelegate(r, uint64(requestID))
	// 	if blockingMode {
	// 		waitGroup.Done()
	// 	}
	// }
	//
	// // Success Callback
	// successCallback := func(envelope *rony.MessageEnvelope) {
	// 	b, _ := envelope.Marshal()
	// 	delegate.OnComplete(b)
	// 	releaseDelegate(r, uint64(requestID))
	// 	if blockingMode {
	// 		waitGroup.Done()
	// 	}
	// }
	//
	// // If this request must be sent to the server then executeRemoteCommand
	// if serverForce {
	// 	executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
	// 	return
	// }
	//
	// // If the constructor is a local command then
	// handler, ok := r.localCommands[constructor]
	// if ok {
	// 	if blockingMode {
	// 		executeLocalCommand(teamID, teamAccess, handler, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
	// 	} else {
	// 		go executeLocalCommand(teamID, teamAccess, handler, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)
	// 	}
	// 	return
	// }
	//
	// // If we reached here, then execute the remote commands
	// executeRemoteCommand(teamID, teamAccess, r, uint64(requestID), constructor, commandBytesDump, timeoutCallback, successCallback)

	return
}
func executeLocalCommand(
	teamID int64, teamAccess uint64,
	handler domain.LocalMessageHandler,
	requestID uint64, constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
) {
	logs.Debug("We execute local command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	in := &rony.MessageEnvelope{
		Header:      domain.TeamHeader(teamID, teamAccess),
		Constructor: constructor,
		Message:     commandBytes,
		RequestID:   requestID,
	}
	out := &rony.MessageEnvelope{
		Header:    domain.TeamHeader(teamID, teamAccess),
		RequestID: requestID,
	}
	handler(in, out, timeoutCB, successCB)
}

// RiverConnection connection info
type RiverConnection struct {
	AuthID  int64
	AuthKey [256]byte
	UserID  int64
}
