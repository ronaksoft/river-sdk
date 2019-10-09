package riversdk

import (
	"context"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_sync"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

var (
	_ServerKeys ServerKeys
)

func SetLogLevel(l int) {
	logs.SetLogLevel(l)
}

// RiverConfig
type RiverConfig struct {
	ServerEndpoint     string
	FileServerEndpoint string
	// QueuePath is the path of a folder that pending requests will be saved there until sending
	// to the server.
	QueuePath string
	// DbPath is the path of the folder holding the sqlite database.
	DbPath string
	// DbID is used to save data for different accounts in separate databases.
	DbID string
	// ServerKeysFilePath is a json file holding finger print and public keys.
	ServerKeys string
	// MainDelegate holds all the general callback functions that let the user of this SDK
	// get notified of the events.
	MainDelegate MainDelegate
	// FileDelegate holds all the callbacks required by file related functions
	FileDelegate FileDelegate

	// LogLevel
	LogLevel int

	// Folder path to save files
	DocumentPhotoDirectory string
	DocumentVideoDirectory string
	DocumentFileDirectory  string
	DocumentAudioDirectory string
	DocumentCacheDirectory string
	DocumentLogDirectory   string
	// Connection Info
	ConnInfo *RiverConnection

	// ClientInfo
	ClientPlatform string
	ClientVersion  string

	// OptimizeForLowMemory if is set then SDK tries to use the lowest possible ram
	OptimizeForLowMemory bool
	MaxInFlightDownloads int32
	MaxInFlightUploads   int32
}

// River
// This the main and a wrapper around all the components of the system (networkController, queueController,
// syncController). All the controllers could be used standalone, but this SDK connect them in a way
// we think is the best possible.
// Only the functions which are exposed will be used by the user of the SDK. All the low-level tasks
// to smooth the connection between client and server are done by this SDK. The underlying storage used
// by this SDK is Sqlite3, however with the least effort you can use other SQL databases. 'model' is the
// package name selected to handle repository functions.
type River struct {
	// Connection Info
	ConnInfo *RiverConnection
	// Device Token
	DeviceToken *msg.AccountRegisterDevice

	// localCommands can be satisfied by client cache
	localCommands map[int64]domain.LocalMessageHandler
	// realTimeCommands should not passed to queue to send they should directly pass to networkController
	realTimeCommands map[int64]bool

	// Internal Controllers
	networkCtrl *networkCtrl.Controller
	queueCtrl   *queueCtrl.Controller
	syncCtrl    *syncCtrl.Controller
	fileCtrl    *fileCtrl.Controller

	// Delegates
	delegateMutex sync.Mutex
	delegates     map[int64]RequestDelegate
	mainDelegate  MainDelegate
	fileDelegate  FileDelegate

	// implements wait 500 ms on out of sync to receive possible missed updates
	lastOutOfSyncTime    time.Time
	chOutOfSyncUpdates   chan []*msg.UpdateContainer
	dbPath               string
	optimizeForLowMemory bool
}

// SetConfig ...
// This function must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	viper.Set(ConfClientPlatform, conf.ClientPlatform)
	viper.Set(ConfClientVersion, conf.ClientVersion)

	r.lastOutOfSyncTime = time.Now().Add(1 * time.Second)
	r.chOutOfSyncUpdates = make(chan []*msg.UpdateContainer, 500)
	r.optimizeForLowMemory = conf.OptimizeForLowMemory
	r.ConnInfo = conf.ConnInfo

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

	r.registerCommandHandlers()
	r.delegates = make(map[int64]RequestDelegate)
	r.mainDelegate = conf.MainDelegate
	r.fileDelegate = conf.FileDelegate

	// set loglevel
	logs.SetLogLevel(conf.LogLevel)

	// set log file path
	if conf.DocumentLogDirectory != "" {
		_ = logs.SetLogFilePath(conf.DocumentLogDirectory)
	}

	// Initialize realtime requests
	r.realTimeCommands = map[int64]bool{
		msg.C_MessagesSetTyping: true,
	}

	// Initialize Network Controller
	r.networkCtrl = networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: conf.ServerEndpoint,
			HttpEndpoint:      conf.FileServerEndpoint,
		},
	)
	r.networkCtrl.SetNetworkStatusChangedCallback(func(newQuality domain.NetworkStatus) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnNetworkStatusChanged(int(newQuality))
		}
	})
	r.networkCtrl.SetErrorHandler(r.onGeneralError)
	r.networkCtrl.SetMessageHandler(r.onReceivedMessage)
	r.networkCtrl.SetUpdateHandler(r.onReceivedUpdate)
	r.networkCtrl.SetOnConnectCallback(r.onNetworkConnect)

	// Initialize FileController
	fileCtrl.SetRootFolders(conf.DocumentAudioDirectory, conf.DocumentFileDirectory, conf.DocumentPhotoDirectory, conf.DocumentVideoDirectory, conf.DocumentCacheDirectory)
	r.fileCtrl = fileCtrl.New(fileCtrl.Config{
		Network:              r.networkCtrl,
		MaxInflightDownloads: conf.MaxInFlightDownloads,
		MaxInflightUploads:   conf.MaxInFlightUploads,
		OnCompleted:          r.fileDelegate.OnCompleted,
		OnProgressChanged:    r.fileDelegate.OnProgressChanged,
		OnCancel:             r.fileDelegate.OnCancel,
		PostUploadProcess:    r.postUploadProcess,
	})

	// Initialize queueController
	if q, err := queueCtrl.New(r.networkCtrl, conf.QueuePath); err != nil {
		logs.Fatal("River::SetConfig() faild to initialize MessageQueue",
			zap.String("Error", err.Error()),
		)
	} else {
		r.queueCtrl = q
	}

	// Initialize Sync Controller
	r.syncCtrl = syncCtrl.NewSyncController(
		syncCtrl.Config{
			ConnInfo:    r.ConnInfo,
			NetworkCtrl: r.networkCtrl,
			QueueCtrl:   r.queueCtrl,
			FileCtrl:    r.fileCtrl,
		},
	)

	// call external delegate on sync status changed
	r.syncCtrl.SetSyncStatusChangedCallback(func(newStatus domain.SyncStatus) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnSyncStatusChanged(int(newStatus))
		}
	})
	// call external delegate on OnUpdate
	r.syncCtrl.SetOnUpdateCallback(func(constructorID int64, b []byte) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnUpdates(constructorID, b)
		}
	})

	// Initialize Server Keys
	if err := _ServerKeys.UnmarshalJSON([]byte(conf.ServerKeys)); err != nil {
		logs.Fatal("River::SetConfig() failed to unmarshal server keys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	logs.Info("River::SetConfig() Load/Create New River Connection")
}

func (r *River) Version() string {
	// TODO:: automatic generation
	return "0.8.1"
}

func (r *River) Start() error {
	logs.Info("River Starting")

	// Initialize MessageHole
	messageHole.Init()

	// Initialize DB replaced with ORM
	repo.InitRepo(r.dbPath, r.optimizeForLowMemory)

	// Update Authorizations
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	r.syncCtrl.SetUserID(r.ConnInfo.UserID)
	r.loadDeviceToken()

	// init UI Executor
	uiexec.InitUIExec()

	// Start Controllers
	r.networkCtrl.Start()
	r.queueCtrl.Start()
	r.syncCtrl.Start()
	r.fileCtrl.Start()

	lastReIndexTime, err := repo.System.LoadInt(domain.SkReIndexTime)
	if err != nil || time.Now().Unix()-int64(lastReIndexTime) > domain.Day {
		go func() {
			logs.Info("ReIndexing Users & Groups")
			repo.Users.ReIndex()
			repo.Groups.ReIndex()
			_ = repo.System.SaveInt(domain.SkReIndexTime, uint64(time.Now().Unix()))
		}()
	}

	logs.Info("River Started")
	return nil
}

func (r *River) Migrate() int {
	ver := r.ConnInfo.Version
	for {
		if f, ok := funcHolders[ver]; ok {
			f(r)
			ver++
		} else {
			if r.ConnInfo.Version != ver {
				r.ConnInfo.Version = ver
				r.ConnInfo.Save()
			}
			return ver
		}
	}
}

func (r *River) onNetworkConnect() {
	// Get Server Time and set server time difference
	timeReq := new(msg.SystemGetServerTime)
	timeReqBytes, _ := timeReq.Marshal()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		for {
			err := r.queueCtrl.ExecuteRealtimeCommand(
				uint64(domain.SequentialUniqueID()),
				msg.C_SystemGetServerTime,
				timeReqBytes,
				nil,
				func(m *msg.MessageEnvelope) {
					switch m.Constructor {
					case msg.C_SystemServerTime:
						x := new(msg.SystemServerTime)
						err := x.Unmarshal(m.Message)
						if err != nil {
							logs.Error("onGetServerTime()", zap.Error(err))
							return
						}
						clientTime := time.Now().Unix()
						serverTime := x.Timestamp
						domain.TimeDelta = time.Duration(serverTime-clientTime) * time.Second

						logs.Debug("River::onGetServerTime()",
							zap.Int64("ServerTime", serverTime),
							zap.Int64("ClientTime", clientTime),
							zap.Duration("Difference", domain.TimeDelta),
						)
					}
				},
				true, false,
			)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(ronak.RandomInt(1000)) * time.Millisecond)
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		r.syncCtrl.UpdateSalt()
		req := msg.AuthRecall{}
		reqBytes, _ := req.Marshal()
		if r.syncCtrl.GetUserID() != 0 {
			// send auth recall until it succeed
			for {
				// this is priority command that should not passed to queue
				// after auth recall answer got back the queue should send its requests in order to get related updates
				err := r.queueCtrl.ExecuteRealtimeCommand(
					uint64(domain.SequentialUniqueID()),
					msg.C_AuthRecall,
					reqBytes,
					nil,
					func(m *msg.MessageEnvelope) {
						if m.Constructor == msg.C_AuthRecalled {
							x := new(msg.AuthRecalled)
							err := x.Unmarshal(m.Message)
							if err != nil {
								logs.Error("onAuthRecalled()", zap.Error(err))
								return
							}
						}
					},
					true,
					false,
				)
				if err == nil {
					break
				}
				time.Sleep(time.Duration(ronak.RandomInt(1000)) * time.Millisecond)
			}
		}
		if r.DeviceToken == nil || r.DeviceToken.Token == "" {
			logs.Warn("onNetworkConnect() Device Token is not set")
		}

	}()

	waitGroup.Wait()

	go func() {
		if r.syncCtrl.GetUserID() != 0 {
			// Sync with Server
			r.syncCtrl.Sync()

			// import contact from server
			r.syncCtrl.ContactImportFromServer()
		}
	}()
}

func (r *River) onGeneralError(e *msg.Error) {
	logs.Info("River::onGeneralError()",
		zap.String("Code", e.Code),
		zap.String("Item", e.Items),
	)
	if e.Code == msg.ErrCodeInvalid && e.Items == msg.ErrItemSalt {
		r.syncCtrl.UpdateSalt()
	}
	if r.mainDelegate != nil {
		buff, _ := e.Marshal()
		r.mainDelegate.OnGeneralError(buff)
	}
}

func (r *River) onReceivedMessage(msgs []*msg.MessageEnvelope) {
	// sort messages by requestID
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].RequestID < msgs[j].RequestID
	})

	// sync localDB with responses in the background
	r.syncCtrl.MessageHandler(msgs)

	// check requestCallbacks and call callbacks
	for idx := range msgs {
		cb := domain.GetRequestCallback(msgs[idx].RequestID)
		if cb == nil {
			continue
		}

		mon.ServerResponseTime(msgs[idx].Constructor, time.Now().Sub(cb.RequestTime))
		select {
		case cb.ResponseChannel <- msgs[idx]:
			logs.Debug("River::onReceivedMessage() passed to callback listener", zap.Uint64("RequestID", cb.RequestID))
		default:
			logs.Error("River::onReceivedMessage() there is no callback listener", zap.Uint64("RequestID", cb.RequestID))
		}
		domain.RemoveRequestCallback(msgs[idx].RequestID)
	}
}

func (r *River) onReceivedUpdate(updateContainer *msg.UpdateContainer) {
	for _, update := range updateContainer.Updates {
		logs.UpdateLog(update.UpdateID, update.Constructor)
	}

	outOfSync := false
	if updateContainer.MinUpdateID != 0 && updateContainer.MinUpdateID > r.syncCtrl.UpdateID()+1 {
		logs.Info("We are out of sync",
			zap.Int64("ContainerMinID", updateContainer.MinUpdateID),
			zap.Int64("ClientUpdateID", r.syncCtrl.UpdateID()),
		)
		outOfSync = true
	}

	r.syncCtrl.UpdateHandler(updateContainer, outOfSync)

	if outOfSync {
		go r.syncCtrl.Sync()
	}
}

func (r *River) postUploadProcess(uploadRequest fileCtrl.UploadRequest) {
	logs.Info("UploadProcess",
		zap.Bool("IsProfile", uploadRequest.IsProfilePhoto),
		zap.Int64("ur.MessageID", uploadRequest.MessageID),
	)
	switch {
	case uploadRequest.IsProfilePhoto == false && uploadRequest.MessageID != 0:
		// This is a upload for message send media
		pendingMessage := repo.PendingMessages.GetByID(uploadRequest.MessageID)
		if pendingMessage == nil {
			return
		}

		repo.PendingMessages.UpdateClientMessageMedia(pendingMessage, uploadRequest.TotalParts)

		req := new(msg.ClientSendMessageMedia)
		_ = req.Unmarshal(pendingMessage.Media)

		// Create SendMessageMedia Request
		x := new(msg.MessagesSendMedia)
		x.Peer = req.Peer
		x.ClearDraft = req.ClearDraft
		x.MediaType = req.MediaType
		x.RandomID = uploadRequest.FileID
		x.ReplyTo = req.ReplyTo

		switch x.MediaType {
		case msg.InputMediaTypeUploadedDocument:
			doc := new(msg.InputMediaUploadedDocument)
			doc.MimeType = req.FileMIME
			doc.Attributes = req.Attributes

			doc.Caption = req.Caption
			doc.File = &msg.InputFile{
				FileID:      uploadRequest.FileID,
				FileName:    req.FileName,
				MD5Checksum: "",
			}

			if uploadRequest.ThumbID != 0 {
				doc.Thumbnail = &msg.InputFile{
					FileID:      uploadRequest.ThumbID,
					FileName:    "thumb_" + req.FileName,
					MD5Checksum: "",
				}
			}

			x.MediaData, _ = doc.Marshal()
		default:

		}
		reqBuff, _ := x.Marshal()
		requestID := uint64(uploadRequest.FileID)
		waitGroup :=sync.WaitGroup{}
		waitGroup.Add(1)
		r.queueCtrl.ExecuteCommand(requestID, msg.C_MessagesSendMedia, reqBuff, nil, func(m *msg.MessageEnvelope) {
			waitGroup.Done()
		}, false)
		waitGroup.Wait()
	case uploadRequest.IsProfilePhoto && uploadRequest.GroupID == 0:
		// This is a upload account profile picture
		x := new(msg.AccountUploadPhoto)
		x.File = &msg.InputFile{
			FileID:      uploadRequest.FileID,
			FileName:    strconv.FormatInt(uploadRequest.FileID, 10) + ".jpg",
			TotalParts:  uploadRequest.TotalParts,
			MD5Checksum: "",
		}
		reqBuff, err := x.Marshal()
		if err != nil {
			logs.Error("SDK::postUploadProcess() marshal AccountUploadPhoto", zap.Error(err))
			return
		}
		requestID := uint64(domain.SequentialUniqueID())
		successCB := func(m *msg.MessageEnvelope) {
			logs.Debug("AccountUploadPhoto success callback")
			switch m.Constructor {
			case msg.C_Bool:
				x := new(msg.Bool)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("AccountUploadPhoto success callback", zap.Error(err))
				}
			case msg.C_Error:
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("AccountUploadPhoto Error callback", zap.Error(err))
				}
				logs.Error("AccountUploadPhoto Error callback", zap.String("Code", x.Code), zap.String("Item", x.Items))
			}
		}
		timeoutCB := func() {
			logs.Debug("AccountUploadPhoto time out callback")
		}
		r.queueCtrl.ExecuteCommand(requestID, msg.C_AccountUploadPhoto, reqBuff, timeoutCB, successCB, false)
	case uploadRequest.IsProfilePhoto && uploadRequest.GroupID != 0:
		// This is a upload group profile picture
		x := new(msg.GroupsUploadPhoto)
		x.GroupID = uploadRequest.GroupID
		x.File = &msg.InputFile{
			FileID:      uploadRequest.FileID,
			FileName:    strconv.FormatInt(uploadRequest.FileID, 10) + ".jpg",
			TotalParts:  uploadRequest.TotalParts,
			MD5Checksum: "",
		}
		reqBuff, err := x.Marshal()
		if err != nil {
			logs.Error("SDK::postUploadProcess() marshal GroupUploadPhoto", zap.Error(err))
			return
		}
		requestID := uint64(domain.SequentialUniqueID())
		successCB := func(m *msg.MessageEnvelope) {
			logs.Debug("GroupUploadPhoto success callback")
			if m.Constructor == msg.C_Bool {
				x := new(msg.Bool)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("GroupUploadPhoto success callback", zap.Error(err))
				}

			}
			if m.Constructor == msg.C_Error {
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("GroupUploadPhoto timeout callback", zap.Error(err))
				}
				logs.Error("GroupUploadPhoto timeout callback", zap.String("Code", x.Code), zap.String("Item", x.Items))
			}
		}
		timeoutCB := func() {
			logs.Debug("GroupUploadPhoto timeoput callback")
		}
		r.queueCtrl.ExecuteCommand(requestID, msg.C_GroupsUploadPhoto, reqBuff, timeoutCB, successCB, false)
	}

}

func (r *River) registerCommandHandlers() {
	r.localCommands = map[int64]domain.LocalMessageHandler{
		msg.C_MessagesGetDialogs:       r.messagesGetDialogs,
		msg.C_MessagesGetDialog:        r.messagesGetDialog,
		msg.C_MessagesGetHistory:       r.messagesGetHistory,
		msg.C_MessagesSend:             r.messagesSend,
		msg.C_ContactsGet:              r.contactsGet,
		msg.C_MessagesReadHistory:      r.messagesReadHistory,
		msg.C_UsersGet:                 r.usersGet,
		msg.C_MessagesGet:              r.messagesGet,
		msg.C_AccountUpdateUsername:    r.accountUpdateUsername,
		msg.C_AccountUpdateProfile:     r.accountUpdateProfile,
		msg.C_AccountRegisterDevice:    r.accountRegisterDevice,
		msg.C_AccountUnregisterDevice:  r.accountUnregisterDevice,
		msg.C_AccountSetNotifySettings: r.accountSetNotifySettings,
		msg.C_MessagesToggleDialogPin:  r.dialogTogglePin,
		msg.C_GroupsEditTitle:          r.groupsEditTitle,
		msg.C_MessagesClearHistory:     r.messagesClearHistory,
		msg.C_MessagesDelete:           r.messagesDelete,
		msg.C_GroupsAddUser:            r.groupAddUser,
		msg.C_GroupsDeleteUser:         r.groupDeleteUser,
		msg.C_GroupsGetFull:            r.groupsGetFull,
		msg.C_GroupsUpdateAdmin:        r.groupUpdateAdmin,
		msg.C_ContactsImport:           r.contactsImport,
		msg.C_MessagesReadContents:     r.messagesReadContents,
		msg.C_UsersGetFull:             r.usersGetFull,
		msg.C_AccountRemovePhoto:       r.accountRemovePhoto,
		msg.C_GroupsRemovePhoto:        r.groupRemovePhoto,
		msg.C_MessagesSendMedia:        r.messagesSendMedia,
		msg.C_ClientSendMessageMedia:   r.clientSendMessageMedia,
		msg.C_MessagesSaveDraft:        r.messagesSaveDraft,
		msg.C_MessagesClearDraft:       r.messagesClearDraft,
	}

}

// GetWorkGroup
// Client call GetWorkGroup with a timeout set, if this function could connect to server and get its response back from
// the server then it returns the serialized version of msg.SystemInfo, otherwise it returns an error
// It is upto the caller to re-call this function in case of error returned.
func GetWorkGroup(url string, timeoutSecond int) ([]byte, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	defer cancelFunc()

	b, err := getWorkGroup(ctx, url)
	return b, err
}

func getWorkGroup(ctx context.Context, url string) ([]byte, error) {
	networkCtrl := networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: url,
		},
	)

	ch := make(chan []byte)
	// Assign Handlers
	// OnConnect Handler
	networkCtrl.SetOnConnectCallback(func() {
		msgEnvelope := new(msg.MessageEnvelope)
		msgEnvelope.RequestID = ronak.RandomUint64()
		msgEnvelope.Constructor = msg.C_SystemGetInfo

		msgEnvelope.Message, _ = (&msg.SystemGetInfo{}).Marshal()
		_ = networkCtrl.SendWebsocket(msgEnvelope, true)
	})
	// Message Handler
	networkCtrl.SetMessageHandler(func(messages []*msg.MessageEnvelope) {
		for _, message := range messages {
			switch message.Constructor {
			case msg.C_SystemInfo:
				ch <- message.Message
				return
			}
		}
	})
	// Update Handler
	networkCtrl.SetUpdateHandler(func(messages *msg.UpdateContainer) {
		// We don't need to handle updates
		return
	})

	// Start the Network Controller alone
	networkCtrl.Start()
	go networkCtrl.Connect()
	defer networkCtrl.Stop()       // 2nd Stop the controller
	defer networkCtrl.Disconnect() // 1st Disconnect

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case b := <-ch:
			return b, nil
		}
	}

}
