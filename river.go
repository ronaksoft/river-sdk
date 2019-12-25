package riversdk

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"go.uber.org/zap"
	"runtime"
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
	ClientOs       string
	ClientVendor   string

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
	domain.ClientPlatform = conf.ClientPlatform
	domain.ClientVersion = conf.ClientVersion
	domain.ClientOS = conf.ClientOs
	domain.ClientVendor = conf.ClientVendor

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
	r.networkCtrl.OnNetworkStatusChange = func(newQuality domain.NetworkStatus) {
		if r.mainDelegate != nil {
			r.mainDelegate.OnNetworkStatusChanged(int(newQuality))
		}
	}
	r.networkCtrl.OnGeneralError = r.onGeneralError
	r.networkCtrl.OnMessage = r.onReceivedMessage
	r.networkCtrl.OnUpdate = r.onReceivedUpdate
	r.networkCtrl.OnWebsocketConnect = r.onNetworkConnect

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
		logs.Fatal("We couldn't initialize MessageQueue",
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
		logs.Error("We couldn't unmarshal ServerKeys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	logs.Info("River SetConfig done!")
}

func (r *River) Version() string {
	return domain.SDKVersion
}

func (r *River) Start() error {
	runtime.GOMAXPROCS(runtime.NumCPU() * 10)
	logs.Info("River Starting")

	// Initialize MessageHole
	messageHole.Init()

	// Initialize DB replaced with ORM
	repo.InitRepo(r.dbPath, r.optimizeForLowMemory)

	// Update Authorizations
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	r.syncCtrl.SetUserID(r.ConnInfo.UserID)
	domain.ClientPhone = r.ConnInfo.Phone
	logs.SetSentry(r.ConnInfo.AuthID, r.ConnInfo.UserID)
	r.loadDeviceToken()

	// init UI Executor
	uiexec.InitUIExec()

	// Update the current salt
	salt.UpdateSalt()

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

	domain.StartTime = time.Now()
	domain.WindowLog = func(txt string) {
		r.mainDelegate.AddLog(txt)
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

func (r *River) onNetworkConnect() (err error) {
	domain.WindowLog(fmt.Sprintf("Connected: %s", domain.StartTime.Format(time.Kitchen)))
	var serverUpdateID int64
	waitGroup := sync.WaitGroup{}
	// If we have no salt then we must call GetServerTime and GetServerSalt sequentially, otherwise
	// We call them in parallel
	err = r.syncCtrl.GetServerTime()
	if err != nil {
		return err
	}
	domain.WindowLog(fmt.Sprintf("ServerTime (%s): %s", domain.TimeDelta, time.Now().Sub(domain.StartTime)))

	if salt.Count() == 0 {
		r.syncCtrl.GetServerSalt()
		domain.WindowLog(fmt.Sprintf("ServerSalt: %s", time.Now().Sub(domain.StartTime)))
	} else {
		if salt.Count() < 3 {
			waitGroup.Add(1)
			go func() {
				r.syncCtrl.GetServerSalt()
				domain.WindowLog(fmt.Sprintf("ServerSalt: %s", time.Now().Sub(domain.StartTime)))
				waitGroup.Done()
			}()
		}
	}
	waitGroup.Add(1)
	go func() {
		// FIXME:: We have server update id here, it is better to call sync only if necessary
		serverUpdateID, err = r.syncCtrl.AuthRecall("NetworkConnect")
		domain.WindowLog(fmt.Sprintf("AuthRecalled: %s", time.Now().Sub(domain.StartTime)))
		waitGroup.Done()
	}()
	waitGroup.Wait()

	if r.networkCtrl.GetQuality() == domain.NetworkDisconnected || err != nil {
		return
	}
	go func() {
		if r.syncCtrl.GetUserID() != 0 {
			if r.syncCtrl.UpdateID() < serverUpdateID {
				// Sync with Server
				r.syncCtrl.Sync()
				domain.WindowLog(fmt.Sprintf("Synced: %s", time.Now().Sub(domain.StartTime)))
			} else if serverUpdateID == 0 {
				r.networkCtrl.Reconnect()
			} else {
				r.syncCtrl.SetSynced()
				domain.WindowLog(fmt.Sprintf("Already Synced: %s", time.Now().Sub(domain.StartTime)))
			}

			// import contact from server
			r.syncCtrl.ContactImportFromServer()
			domain.WindowLog(fmt.Sprintf("ContactsImported: %s", time.Now().Sub(domain.StartTime)))
		}
	}()
	return nil
}

func (r *River) onGeneralError(requestID uint64, e *msg.Error) {
	logs.Info("We received error (General)",
		zap.Uint64("RequestID", requestID),
		zap.String("Code", e.Code),
		zap.String("Item", e.Items),
	)
	if e.Code == msg.ErrCodeInvalid && e.Items == msg.ErrItemSalt {
		if !salt.UpdateSalt() {
			go func() {
				r.syncCtrl.GetServerSalt()
				domain.WindowLog(fmt.Sprintf("SaltsReceived: %s", time.Now().Sub(domain.StartTime)))
			}()
		}
	}
	if r.mainDelegate != nil && requestID == 0 {
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
			logs.Debug("We received response",
				zap.Uint64("ReqID", cb.RequestID),
				zap.String("Constructor", msg.ConstructorNames[msgs[idx].Constructor]),
			)
		default:
			logs.Error("We received response but no callback, we drop response",
				zap.Uint64("ReqID", cb.RequestID),
				zap.String("Constructor", msg.ConstructorNames[msgs[idx].Constructor]),
			)
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
	logs.Info("Upload finished, we process the next action",
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
		waitGroup := sync.WaitGroup{}
		waitGroup.Add(1)
		r.queueCtrl.EnqueueCommand(requestID, msg.C_MessagesSendMedia, reqBuff, nil, func(m *msg.MessageEnvelope) {
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
			logs.Error("We couldn't marshal AccountUploadPhoto", zap.Error(err))
			return
		}
		requestID := uint64(domain.SequentialUniqueID())
		successCB := func(m *msg.MessageEnvelope) {
			logs.Debug("AccountUploadPhoto success callback called")
			switch m.Constructor {
			case msg.C_Bool:
				x := new(msg.Bool)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("We couldn't unmarshal AccountUploadPhoto (Bool) response", zap.Error(err))
				}
			case msg.C_Error:
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("We couldn't unmarshal AccountUploadPhoto (Error) response", zap.Error(err))
				}
				logs.Error("We received error on AccountUploadPhoto response", zap.String("Code", x.Code), zap.String("Item", x.Items))
			}
		}
		timeoutCB := func() {
			logs.Debug("Timeout! on AccountUploadPhoto response")
		}
		r.queueCtrl.EnqueueCommand(requestID, msg.C_AccountUploadPhoto, reqBuff, timeoutCB, successCB, false)
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
			logs.Error("We couldn't marshal GroupUploadPhoto", zap.Error(err))
			return
		}
		requestID := uint64(domain.SequentialUniqueID())
		successCB := func(m *msg.MessageEnvelope) {
			logs.Debug("GroupUploadPhoto success callback called")
			if m.Constructor == msg.C_Bool {
				x := new(msg.Bool)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("We couldn't unmarshal GroupUploadPhoto (Bool) response", zap.Error(err))
				}

			}
			if m.Constructor == msg.C_Error {
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("We couldn't unmarshal GroupUploadPhoto (Error) response", zap.Error(err))
				}
				logs.Error("We received error on GroupUploadPhoto response", zap.String("Code", x.Code), zap.String("Item", x.Items))
			}
		}
		timeoutCB := func() {
			logs.Debug("GTimeout! on GroupUploadPhoto response")
		}
		r.queueCtrl.EnqueueCommand(requestID, msg.C_GroupsUploadPhoto, reqBuff, timeoutCB, successCB, false)
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
		msg.C_LabelsListItems:          r.labelsListItems,
		msg.C_ClientGlobalSearch:       r.clientGlobalSearch,
		msg.C_ClientContactSearch:      r.clientContactSearch,
	}
}
