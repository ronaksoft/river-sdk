package riversdk

import (
	"fmt"
	msg "git.ronaksoftware.com/river/msg/chat"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
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
	CountryCode    string

	// OptimizeForLowMemory if is set then SDK tries to use the lowest possible ram
	OptimizeForLowMemory bool
	MaxInFlightDownloads int32
	MaxInFlightUploads   int32

	// Misc
	ResetQueueOnStartup bool
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
	delegates     map[uint64]RequestDelegate
	mainDelegate  MainDelegate
	fileDelegate  FileDelegate

	dbPath               string
	optimizeForLowMemory bool
	resetQueueOnStartup  bool
}

// SetConfig ...
// This function must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	domain.ClientPlatform = conf.ClientPlatform
	domain.ClientVersion = conf.ClientVersion
	domain.ClientOS = conf.ClientOs
	domain.ClientVendor = conf.ClientVendor

	r.optimizeForLowMemory = conf.OptimizeForLowMemory
	r.resetQueueOnStartup = conf.ResetQueueOnStartup
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
	r.delegates = make(map[uint64]RequestDelegate)
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
			CountryCode:       conf.CountryCode,
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
	repo.Files.SetRootFolders(
		conf.DocumentAudioDirectory,
		conf.DocumentFileDirectory,
		conf.DocumentPhotoDirectory,
		conf.DocumentVideoDirectory,
		conf.DocumentCacheDirectory,
	)
	r.fileCtrl = fileCtrl.New(fileCtrl.Config{
		Network:              r.networkCtrl,
		MaxInflightDownloads: conf.MaxInFlightDownloads,
		MaxInflightUploads:   conf.MaxInFlightUploads,
		HttpRequestTimeout:   domain.HttpRequestTime,
		CompletedCB:          r.fileDelegate.OnCompleted,
		ProgressChangedCB:    r.fileDelegate.OnProgressChanged,
		CancelCB:             r.fileDelegate.OnCancel,
		PostUploadProcessCB:  r.postUploadProcess,
	})

	// Initialize queueController
	if q, err := queueCtrl.New(r.fileCtrl, r.networkCtrl, r.dbPath); err != nil {
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
			SyncStatusChangeCB: func(newStatus domain.SyncStatus) {
				if r.mainDelegate != nil {
					r.mainDelegate.OnSyncStatusChanged(int(newStatus))
				}
			},
			UpdateReceivedCB: func(constructorID int64, b []byte) {
				if r.mainDelegate != nil {
					r.mainDelegate.OnUpdates(constructorID, b)
				}
			},
			AppUpdateCB: func(version string, updateAvailable bool, force bool) {
				if r.mainDelegate != nil {
					r.mainDelegate.AppUpdate(version, updateAvailable, force)
				}
			},
		},
	)

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
		serverUpdateID, err = r.syncCtrl.AuthRecall("NetworkConnect")
		if err != nil {
			logs.Warn("Error On AuthRecall", zap.Error(err))
		}
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
			} else {
				r.syncCtrl.SetSynced()
				domain.WindowLog(fmt.Sprintf("Already Synced: %s", time.Now().Sub(domain.StartTime)))
			}

			_, err := repo.System.LoadBytes("SysConfig")
			if err != nil {
				r.syncCtrl.GetSystemConfig()
			}

			// Get contacts and imports remaining contacts
			r.syncCtrl.ContactsGet()
			domain.WindowLog(fmt.Sprintf("ContactsGet: %s", time.Now().Sub(domain.StartTime)))
			r.syncCtrl.ContactsImport(true, nil, nil)
			domain.WindowLog(fmt.Sprintf("ContactsImported: %s", time.Now().Sub(domain.StartTime)))
		}
	}()
	return nil
}

func (r *River) onGeneralError(requestID uint64, e *msg.Error) {
	logs.Info("We received error (General)",
		zap.Uint64("ReqID", requestID),
		zap.String("Code", e.Code),
		zap.String("Item", e.Items),
	)
	switch {
	case e.Code == msg.ErrCodeInvalid && e.Items == msg.ErrItemSalt:
		if !salt.UpdateSalt() {
			go func() {
				r.syncCtrl.GetServerSalt()
				domain.WindowLog(fmt.Sprintf("SaltsReceived: %s", time.Now().Sub(domain.StartTime)))
			}()
		}
	case e.Code == msg.ErrCodeUnavailable && e.Items == msg.ErrItemUserID:
		go r.Logout(false, 1)
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
		reqCB := domain.GetRequestCallback(msgs[idx].RequestID)
		if reqCB == nil {
			continue
		}

		mon.ServerResponseTime(reqCB.Constructor, msgs[idx].Constructor, time.Now().Sub(reqCB.DepartureTime))
		select {
		case reqCB.ResponseChannel <- msgs[idx]:
			logs.Debug("We received response",
				zap.Uint64("ReqID", reqCB.RequestID),
				zap.String("C", msg.ConstructorNames[msgs[idx].Constructor]),
			)
		default:
			logs.Error("We received response but no callback, we drop response",
				zap.Uint64("ReqID", reqCB.RequestID),
				zap.String("C", msg.ConstructorNames[msgs[idx].Constructor]),
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

func (r *River) postUploadProcess(uploadRequest fileCtrl.UploadRequest) bool {
	logs.Info("Upload finished, we process the next action",
		zap.Bool("IsProfile", uploadRequest.IsProfilePhoto),
		zap.Int64("ur.MessageID", uploadRequest.MessageID),
	)
	switch {
	case uploadRequest.IsProfilePhoto == false && uploadRequest.MessageID != 0:
		return r.sendMessageMedia(uploadRequest)
	case uploadRequest.IsProfilePhoto && uploadRequest.GroupID == 0:
		return r.uploadAccountPhoto(uploadRequest)
	case uploadRequest.IsProfilePhoto && uploadRequest.GroupID != 0:
		return r.uploadGroupPhoto(uploadRequest)
	}
	return false
}
func (r *River) sendMessageMedia(uploadRequest fileCtrl.UploadRequest) (success bool) {
	// This is a upload for message send media
	pendingMessage := repo.PendingMessages.GetByID(uploadRequest.MessageID)
	if pendingMessage == nil {
		return true
	}

	req := &msg.ClientSendMessageMedia{}
	_ = req.Unmarshal(pendingMessage.Media)

	err := ronak.Try(3, time.Millisecond*500, func() error {
		var fileLoc *msg.FileLocation
		if uploadRequest.DocumentID != 0 && uploadRequest.AccessHash != 0 && uploadRequest.ClusterID != 0 {
			req.MediaType = msg.InputMediaTypeDocument
			fileLoc = &msg.FileLocation{
				ClusterID:  uploadRequest.ClusterID,
				FileID:     uploadRequest.DocumentID,
				AccessHash: uploadRequest.AccessHash,
			}
		}
		return repo.PendingMessages.UpdateClientMessageMedia(pendingMessage, uploadRequest.TotalParts, req.MediaType, fileLoc)
	})
	if err != nil {
		logs.Error("Error On UpdateClientMessageMedia", zap.Error(err))
	}

	// Create SendMessageMedia Request
	x := &msg.MessagesSendMedia{
		Peer:       req.Peer,
		ClearDraft: req.ClearDraft,
		MediaType:  req.MediaType,
		RandomID:   uploadRequest.FileID,
		ReplyTo:    req.ReplyTo,
	}

	switch x.MediaType {
	case msg.InputMediaTypeUploadedDocument:
		doc := &msg.InputMediaUploadedDocument{
			MimeType:   req.FileMIME,
			Attributes: req.Attributes,
			Caption:    req.Caption,
			Entities:   req.Entities,
			File: &msg.InputFile{
				FileID:      uploadRequest.FileID,
				FileName:    req.FileName,
				MD5Checksum: "",
			},
		}
		if uploadRequest.ThumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      uploadRequest.ThumbID,
				FileName:    "thumb_" + req.FileName,
				MD5Checksum: "",
			}
		}
		x.MediaData, _ = doc.Marshal()
	case msg.InputMediaTypeDocument:
		doc := &msg.InputMediaDocument{
			Caption:    req.Caption,
			Attributes: req.Attributes,
			Entities:   req.Entities,
			Document: &msg.InputDocument{
				ID:         uploadRequest.DocumentID,
				AccessHash: uploadRequest.AccessHash,
				ClusterID:  uploadRequest.ClusterID,
			},
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
	success = true

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	successCB := func(m *msg.MessageEnvelope) {
		logs.Debug("MessagesSendMedia success callback called")
		switch m.Constructor {
		case msg.C_Error:
			success = false
			x := &msg.Error{}
			if err := x.Unmarshal(m.Message); err != nil {
				logs.Error("We couldn't unmarshal MessagesSendMedia (Error) response", zap.Error(err))
			}
			logs.Error("We received error on MessagesSendMedia response",
				zap.String("Code", x.Code),
				zap.String("Item", x.Items),
			)
			if x.Code == msg.ErrCodeAlreadyExists && x.Items == msg.ErrItemRandomID {
				success = true
				_ = repo.PendingMessages.Delete(uploadRequest.MessageID)

			}
		}
		waitGroup.Done()
	}
	timeoutCB := func() {
		success = false
		logs.Debug("We got Timeout! on MessagesSendMedia response")
		waitGroup.Done()
	}
	r.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(x.RandomID),
			Message:     reqBuff,
		},
		timeoutCB, successCB, false)
	waitGroup.Wait()
	return
}
func (r *River) uploadGroupPhoto(uploadRequest fileCtrl.UploadRequest) (success bool) {
	// This is a upload group profile picture
	x := &msg.GroupsUploadPhoto{
		GroupID: uploadRequest.GroupID,
		File: &msg.InputFile{
			FileID:      uploadRequest.FileID,
			FileName:    strconv.FormatInt(uploadRequest.FileID, 10) + ".jpg",
			TotalParts:  uploadRequest.TotalParts,
			MD5Checksum: "",
		},
	}

	reqBuff, _ := x.Marshal()

	success = true
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	successCB := func(m *msg.MessageEnvelope) {
		logs.Debug("GroupUploadPhoto success callback called")
		switch m.Constructor {
		case msg.C_Error:
			success = false
			x := &msg.Error{}
			if err := x.Unmarshal(m.Message); err != nil {
				logs.Error("We couldn't unmarshal GroupUploadPhoto (Error) response", zap.Error(err))
			}
			logs.Error("We received error on GroupUploadPhoto response", zap.String("Code", x.Code), zap.String("Item", x.Items))
		}
		waitGroup.Done()
	}
	timeoutCB := func() {
		success = false
		logs.Debug("We got Timeout! on GroupUploadPhoto response")
		waitGroup.Done()
	}
	r.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_GroupsUploadPhoto,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBuff,
		},
		timeoutCB, successCB, false,
	)
	waitGroup.Wait()
	return
}
func (r *River) uploadAccountPhoto(uploadRequest fileCtrl.UploadRequest) (success bool) {
	// This is a upload account profile picture
	x := &msg.AccountUploadPhoto{
		File: &msg.InputFile{
			FileID:      uploadRequest.FileID,
			FileName:    strconv.FormatInt(uploadRequest.FileID, 10) + ".jpg",
			TotalParts:  uploadRequest.TotalParts,
			MD5Checksum: "",
		},
	}
	reqBuff, _ := x.Marshal()
	success = true
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	successCB := func(m *msg.MessageEnvelope) {
		logs.Debug("AccountUploadPhoto success callback called")
		switch m.Constructor {
		case msg.C_Error:
			success = false
			x := &msg.Error{}
			if err := x.Unmarshal(m.Message); err != nil {
				logs.Error("We couldn't unmarshal AccountUploadPhoto (Error) response", zap.Error(err))
			}
			logs.Error("We received error on AccountUploadPhoto response", zap.String("Code", x.Code), zap.String("Item", x.Items))
		}
		waitGroup.Done()
	}
	timeoutCB := func() {
		success = false
		logs.Debug("Timeout! on AccountUploadPhoto response")
		waitGroup.Done()
	}
	r.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_AccountUploadPhoto,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBuff,
		},
		timeoutCB, successCB, false,
	)
	waitGroup.Wait()
	return
}

func (r *River) registerCommandHandlers() {
	r.localCommands = map[int64]domain.LocalMessageHandler{
		msg.C_MessagesGetDialogs:            r.messagesGetDialogs,
		msg.C_MessagesGetDialog:             r.messagesGetDialog,
		msg.C_MessagesGetHistory:            r.messagesGetHistory,
		msg.C_MessagesSend:                  r.messagesSend,
		msg.C_ContactsGet:                   r.contactsGet,
		msg.C_MessagesReadHistory:           r.messagesReadHistory,
		msg.C_UsersGet:                      r.usersGet,
		msg.C_MessagesGet:                   r.messagesGet,
		msg.C_AccountUpdateUsername:         r.accountUpdateUsername,
		msg.C_AccountUpdateProfile:          r.accountUpdateProfile,
		msg.C_AccountRegisterDevice:         r.accountRegisterDevice,
		msg.C_AccountUnregisterDevice:       r.accountUnregisterDevice,
		msg.C_AccountSetNotifySettings:      r.accountSetNotifySettings,
		msg.C_MessagesToggleDialogPin:       r.dialogTogglePin,
		msg.C_GroupsEditTitle:               r.groupsEditTitle,
		msg.C_MessagesClearHistory:          r.messagesClearHistory,
		msg.C_MessagesDelete:                r.messagesDelete,
		msg.C_GroupsAddUser:                 r.groupAddUser,
		msg.C_GroupsDeleteUser:              r.groupDeleteUser,
		msg.C_GroupsGetFull:                 r.groupsGetFull,
		msg.C_GroupsUpdateAdmin:             r.groupUpdateAdmin,
		msg.C_ContactsImport:                r.contactsImport,
		msg.C_ContactsDelete:                r.contactsDelete,
		msg.C_ContactsGetTopPeers:           r.contactsGetTopPeers,
		msg.C_MessagesReadContents:          r.messagesReadContents,
		msg.C_UsersGetFull:                  r.usersGetFull,
		msg.C_AccountRemovePhoto:            r.accountRemovePhoto,
		msg.C_GroupsRemovePhoto:             r.groupRemovePhoto,
		msg.C_MessagesSendMedia:             r.messagesSendMedia,
		msg.C_ClientSendMessageMedia:        r.clientSendMessageMedia,
		msg.C_MessagesSaveDraft:             r.messagesSaveDraft,
		msg.C_MessagesClearDraft:            r.messagesClearDraft,
		msg.C_LabelsListItems:               r.labelsListItems,
		msg.C_ClientGlobalSearch:            r.clientGlobalSearch,
		msg.C_ClientContactSearch:           r.clientContactSearch,
		msg.C_ClientGetCachedMedia:          r.clientGetCachedMedia,
		msg.C_ClientClearCachedMedia:        r.clientClearCachedMedia,
		msg.C_ClientGetMediaHistory:         r.clientGetMediaHistory,
		msg.C_ClientGetLastBotKeyboard:      r.clientGetLastBotKeyboard,
		msg.C_ClientGetRecentSearch:         r.clientGetRecentSearch,
		msg.C_ClientPutRecentSearch:         r.clientPutRecentSearch,
		msg.C_ClientRemoveAllRecentSearches: r.clientRemoveAllRecentSearches,
	}
}
