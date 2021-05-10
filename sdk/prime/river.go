package riversdk

import (
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/module"
	"git.ronaksoft.com/river/sdk/internal/module/account"
	"git.ronaksoft.com/river/sdk/internal/module/contact"
	"git.ronaksoft.com/river/sdk/internal/module/gif"
	"git.ronaksoft.com/river/sdk/internal/module/group"
	"git.ronaksoft.com/river/sdk/internal/module/label"
	"git.ronaksoft.com/river/sdk/internal/module/message"
	"git.ronaksoft.com/river/sdk/internal/module/search"
	"git.ronaksoft.com/river/sdk/internal/module/system"
	"git.ronaksoft.com/river/sdk/internal/module/team"
	"git.ronaksoft.com/river/sdk/internal/module/user"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/salt"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	"git.ronaksoft.com/river/sdk/internal/ctrl_sync"
	"git.ronaksoft.com/river/sdk/internal/domain"
)

func SetLogLevel(l int) {
	logs.SetLogLevel(l)
}

type RiverConfig struct {
	ServerHostPort string
	// DbPath is the path of the folder holding the sqlite database.
	DbPath string
	// DbID is used to save data for different accounts in separate databases. Could be used for multi-account cases.
	DbID string
	// MainDelegate holds all the general callback functions that let the user of this SDK
	// get notified of the events.
	MainDelegate MainDelegate
	// FileDelegate holds all the callbacks required by file related functions
	FileDelegate FileDelegate
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

	// OptimizeForLowMemory if is set then SDK tries to use the lowest possible ram
	OptimizeForLowMemory bool
	MaxInFlightDownloads int32
	MaxInFlightUploads   int32

	// Misc
	ResetQueueOnStartup bool

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
	// Connection Info
	ConnInfo *RiverConnection

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
	sentryDSN            string
}

func (r *River) GetConnInfo() domain.RiverConfigurator {
	return r.ConnInfo
}

func (r *River) SyncCtrl() *syncCtrl.Controller {
	return r.syncCtrl
}

func (r *River) NetCtrl() *networkCtrl.Controller {
	return r.networkCtrl
}

func (r *River) QueueCtrl() *queueCtrl.Controller {
	return r.queueCtrl
}

func (r *River) FileCtrl() *fileCtrl.Controller {
	return r.fileCtrl
}

// SetConfig must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	domain.ClientPlatform = conf.ClientPlatform
	domain.ClientVersion = conf.ClientVersion
	domain.ClientOS = conf.ClientOs
	domain.ClientVendor = conf.ClientVendor

	r.sentryDSN = conf.SentryDSN
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

	r.registerModule(
		account.New(),
		contact.New(),
		gif.New(),
		group.New(),
		label.New(),
		message.New(),
		search.New(),
		system.New(),
		team.New(),
		user.New(),
	)

	r.delegates = make(map[uint64]RequestDelegate)
	r.mainDelegate = conf.MainDelegate
	r.fileDelegate = conf.FileDelegate

	// set loglevel
	logs.SetLogLevel(conf.LogLevel)

	// set log file path
	if conf.LogDirectory != "" {
		_ = logs.SetLogFilePath(conf.LogDirectory)
	}

	// Initialize realtime requests
	r.realTimeCommands = map[int64]bool{
		msg.C_MessagesSetTyping:   true,
		msg.C_InitConnect:         true,
		msg.C_InitConnectTest:     true,
		msg.C_InitAuthCompleted:   true,
		msg.C_SystemGetConfig:     true,
		msg.C_SystemGetSalts:      true,
		msg.C_SystemGetServerTime: true,
		msg.C_SystemGetServerKeys: true,
	}

	// Initialize Network Controller
	r.networkCtrl = networkCtrl.New(
		networkCtrl.Config{
			WebsocketEndpoint: fmt.Sprintf("ws://%s", conf.ServerHostPort),
			HttpEndpoint:      fmt.Sprintf("http://%s", conf.ServerHostPort),
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
	repo.SetRootFolders(
		conf.DocumentAudioDirectory,
		conf.DocumentFileDirectory,
		conf.DocumentPhotoDirectory,
		conf.DocumentVideoDirectory,
		conf.DocumentCacheDirectory,
	)
	r.fileCtrl = fileCtrl.New(fileCtrl.Config{
		Network:              r.networkCtrl,
		DbPath:               r.dbPath,
		MaxInflightDownloads: conf.MaxInFlightDownloads,
		MaxInflightUploads:   conf.MaxInFlightUploads,
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

	// Initialize River Connection
	logs.Info("River SetConfig done!")

	// Set current team
	domain.SetCurrentTeam(conf.TeamID, uint64(conf.TeamAccessHash))
}

func (r *River) onNetworkConnect() (err error) {
	defer logs.RecoverPanic(
		"River::onNetworkConnect",
		domain.M{
			"OS":  domain.ClientOS,
			"Ver": domain.ClientVersion,
		},
		nil,
	)

	domain.WindowLog(fmt.Sprintf("Connected: %s", domain.StartTime.Format(time.Kitchen)))
	var serverUpdateID int64
	waitGroup := &sync.WaitGroup{}
	// If we have no salt then we must call GetServerTime and GetServerSalt sequentially, otherwise
	// We call them in parallel
	if atomic.LoadInt32(&domain.TimeSynced) == 0 {
		err = r.syncCtrl.GetServerTime()
		if err != nil {
			return err
		}
		domain.WindowLog(fmt.Sprintf("ServerTime (%s): %s", domain.TimeDelta, time.Now().Sub(domain.StartTime)))
	}
	atomic.CompareAndSwapInt32(&domain.TimeSynced, 0, 1)

	switch salt.Count() {
	case 0:
		r.syncCtrl.GetServerSalt()
		domain.WindowLog(fmt.Sprintf("ServerSalt: %s", time.Now().Sub(domain.StartTime)))
	case 1, 2, 3:
		waitGroup.Add(1)
		go func() {
			r.syncCtrl.GetServerSalt()
			domain.WindowLog(fmt.Sprintf("ServerSalt: %s", time.Now().Sub(domain.StartTime)))
			waitGroup.Done()
		}()
	default:
		// We have enough salts
	}

	serverUpdateID, err = r.syncCtrl.AuthRecall("NetworkConnect")
	if err != nil {
		logs.Warn("Error On AuthRecall", zap.Error(err))
	}
	domain.WindowLog(fmt.Sprintf("AuthRecalled: %s", time.Now().Sub(domain.StartTime)))
	waitGroup.Wait()

	// If we are disconnected or not logged in or error happened then we return
	if err != nil || r.syncCtrl.GetUserID() == 0 || r.networkCtrl.GetQuality() == domain.NetworkDisconnected {
		return
	}

	go func() {
		// Check if client is synced with servers
		if r.syncCtrl.GetUpdateID() < serverUpdateID {
			// Sync with Server
			r.syncCtrl.Sync()
			domain.WindowLog(fmt.Sprintf("Synced: %s", time.Now().Sub(domain.StartTime)))
		} else {
			r.syncCtrl.SetSynced()
			domain.WindowLog(fmt.Sprintf("Already Synced: %s", time.Now().Sub(domain.StartTime)))
		}

		// Load SystemConfigs
		if atomic.LoadInt32(&domain.ConfigSynced) == 0 {
			r.syncCtrl.GetSystemConfig()
		}
		atomic.CompareAndSwapInt32(&domain.ConfigSynced, 0, 1)

		if atomic.LoadInt32(&domain.ContactsSynced) == 0 {
			// Get contacts and imports remaining contacts
			waitGroup.Add(1)
			r.syncCtrl.GetContacts(waitGroup, 0, 0)
			waitGroup.Wait()
			domain.WindowLog(fmt.Sprintf("ContactsGet: %s", time.Now().Sub(domain.StartTime)))
			r.syncCtrl.ContactsImport(true, nil, nil)
			domain.WindowLog(fmt.Sprintf("ContactsImported: %s", time.Now().Sub(domain.StartTime)))
		}
		atomic.CompareAndSwapInt32(&domain.ContactsSynced, 0, 1)

	}()
	return nil
}

func (r *River) onGeneralError(requestID uint64, e *rony.Error) {
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
		// We don't do anything just log, but client must call logout
	}

	if r.mainDelegate != nil && requestID == 0 {
		buff, _ := e.Marshal()
		r.mainDelegate.OnGeneralError(buff)
	}
}

func (r *River) onReceivedMessage(msgs []*rony.MessageEnvelope) {
	defer logs.RecoverPanic(
		"River::onReceivedMessage",
		domain.M{
			"OS":  domain.ClientOS,
			"Ver": domain.ClientVersion,
		},
		nil,
	)

	// sync localDB with responses in the background
	r.syncCtrl.MessageApplier(msgs)

	// check requestCallbacks and call callbacks
	for idx := range msgs {
		reqCB := domain.GetRequestCallback(msgs[idx].RequestID)
		if reqCB == nil {
			continue
		}

		mon.ServerResponseTime(reqCB.Constructor, msgs[idx].Constructor, time.Duration(tools.NanoTime()-reqCB.DepartureTime))
		select {
		case reqCB.ResponseChannel <- msgs[idx]:
			logs.Debug("We received response",
				zap.Uint64("ReqID", reqCB.RequestID),
				zap.String("C", registry.ConstructorName(msgs[idx].Constructor)),
			)
		default:
			logs.Error("We received response but no callback, we drop response",
				zap.Uint64("ReqID", reqCB.RequestID),
				zap.String("C", registry.ConstructorName(msgs[idx].Constructor)),
			)
		}
		domain.RemoveRequestCallback(msgs[idx].RequestID)
	}
}

func (r *River) onReceivedUpdate(updateContainer *msg.UpdateContainer) {
	defer logs.RecoverPanic(
		"River::onReceivedUpdate",
		domain.M{
			"OS":              domain.ClientOS,
			"Ver":             domain.ClientVersion,
			"UpdateContainer": updateContainer,
		},
		nil,
	)

	for _, update := range updateContainer.Updates {
		logs.UpdateLog(update.UpdateID, update.Constructor)
	}

	outOfSync := false
	if updateContainer.MinUpdateID != 0 && updateContainer.MinUpdateID > r.syncCtrl.GetUpdateID()+1 {
		logs.Info("We are out of sync",
			zap.Int64("ContainerMinID", updateContainer.MinUpdateID),
			zap.Int64("ClientUpdateID", r.syncCtrl.GetUpdateID()),
		)
		outOfSync = true
	}

	r.syncCtrl.UpdateApplier(updateContainer, outOfSync)

	if outOfSync {
		go r.syncCtrl.Sync()
	}
}

func (r *River) postUploadProcess(uploadRequest *msg.ClientFileRequest) bool {
	defer logs.RecoverPanic(
		"River::postUploadProcess",
		domain.M{
			"OS":       domain.ClientOS,
			"Ver":      domain.ClientVersion,
			"FilePath": uploadRequest.FilePath,
		},
		nil,
	)

	logs.Info("River Post Upload Process",
		zap.Bool("IsProfile", uploadRequest.IsProfilePhoto),
		zap.Int64("MessageID", uploadRequest.MessageID),
		zap.Int64("FileID", uploadRequest.FileID),
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
func (r *River) sendMessageMedia(uploadRequest *msg.ClientFileRequest) (success bool) {
	// This is a upload for message send
	pendingMessage := repo.PendingMessages.GetByID(uploadRequest.MessageID)
	if pendingMessage == nil {
		return true
	}

	req := &msg.ClientSendMessageMedia{}
	_ = req.Unmarshal(pendingMessage.Media)
	err := tools.Try(3, time.Millisecond*500, func() error {
		var fileLoc *msg.FileLocation
		if uploadRequest.FileID != 0 && uploadRequest.AccessHash != 0 && uploadRequest.ClusterID != 0 {
			req.MediaType = msg.InputMediaType_InputMediaTypeDocument
			fileLoc = &msg.FileLocation{
				ClusterID:  uploadRequest.ClusterID,
				FileID:     uploadRequest.FileID,
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
		RandomID:   pendingMessage.FileID,
		ReplyTo:    req.ReplyTo,
	}

	switch x.MediaType {
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
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
			TinyThumbnail: req.TinyThumb,
		}
		if uploadRequest.ThumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      uploadRequest.ThumbID,
				FileName:    "thumb_" + req.FileName,
				MD5Checksum: "",
			}
		}
		x.MediaData, _ = doc.Marshal()
	case msg.InputMediaType_InputMediaTypeDocument:
		doc := &msg.InputMediaDocument{
			Caption:    req.Caption,
			Attributes: req.Attributes,
			Entities:   req.Entities,
			Document: &msg.InputDocument{
				ID:         uploadRequest.FileID,
				AccessHash: uploadRequest.AccessHash,
				ClusterID:  uploadRequest.ClusterID,
			},
			TinyThumbnail: req.TinyThumb,
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
	successCB := func(m *rony.MessageEnvelope) {
		logs.Info("MessagesSendMedia success callback called", zap.String("C", registry.ConstructorName(m.Constructor)))
		switch m.Constructor {
		case rony.C_Error:
			success = false
			x := &rony.Error{}
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
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(x.RandomID),
			Message:     reqBuff,
			Header:      domain.TeamHeader(pendingMessage.TeamID, pendingMessage.TeamAccessHash),
		},
		timeoutCB, successCB, false)
	waitGroup.Wait()
	return
}
func (r *River) uploadGroupPhoto(uploadRequest *msg.ClientFileRequest) (success bool) {
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
	successCB := func(m *rony.MessageEnvelope) {
		logs.Debug("GroupUploadPhoto success callback called")
		switch m.Constructor {
		case rony.C_Error:
			success = false
			x := &rony.Error{}
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
		&rony.MessageEnvelope{
			Constructor: msg.C_GroupsUploadPhoto,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBuff,
		},
		timeoutCB, successCB, false,
	)
	waitGroup.Wait()
	return
}
func (r *River) uploadAccountPhoto(uploadRequest *msg.ClientFileRequest) (success bool) {
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
	successCB := func(m *rony.MessageEnvelope) {
		logs.Debug("AccountUploadPhoto success callback called")
		switch m.Constructor {
		case rony.C_Error:
			success = false
			x := &rony.Error{}
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
		&rony.MessageEnvelope{
			Constructor: msg.C_AccountUploadPhoto,
			RequestID:   uint64(domain.SequentialUniqueID()),
			Message:     reqBuff,
		},
		timeoutCB, successCB, false,
	)
	waitGroup.Wait()
	return
}

func (r *River) registerModule(modules ...module.Module) {
	for _, m := range modules {
		m.Init(r)
		for c, h := range m.LocalHandlers() {
			r.localCommands[c] = h
		}
	}
}

// RiverConnection connection info
type RiverConnection struct {
	AuthID    int64
	AuthKey   [256]byte
	UserID    int64
	Username  string
	Phone     string
	FirstName string
	LastName  string
	Bio       string
	Delegate  ConnInfoDelegate `json:"-"`
	Version   int
}

// Save RiverConfig interface func
func (v *RiverConnection) Save() {
	logs.Debug("ConnInfo saved.")
	b, _ := json.Marshal(v)
	v.Delegate.SaveConnInfo(b)
}

func (v *RiverConnection) ChangeAuthID(authID int64) { v.AuthID = authID }

func (v *RiverConnection) ChangeAuthKey(authKey []byte) {
	copy(v.AuthKey[:], authKey[:256])
}

func (v *RiverConnection) GetAuthKey() []byte {
	var bytes = make([]byte, 256)
	copy(bytes, v.AuthKey[0:256])
	return bytes
}

func (v *RiverConnection) ChangeUserID(userID int64) { v.UserID = userID }

func (v *RiverConnection) ChangeUsername(username string) { v.Username = username }

func (v *RiverConnection) ChangePhone(phone string) {
	v.Phone = phone
}

func (v *RiverConnection) ChangeFirstName(firstName string) { v.FirstName = firstName }

func (v *RiverConnection) ChangeLastName(lastName string) { v.LastName = lastName }

func (v *RiverConnection) ChangeBio(bio string) { v.Bio = bio }

func (v *RiverConnection) PickupAuthID() int64 { return v.AuthID }

func (v *RiverConnection) PickupAuthKey() [256]byte { return v.AuthKey }

func (v *RiverConnection) PickupUserID() int64 { return v.UserID }

func (v *RiverConnection) PickupUsername() string { return v.Username }

func (v *RiverConnection) PickupPhone() string { return v.Phone }

func (v *RiverConnection) PickupFirstName() string { return v.FirstName }

func (v *RiverConnection) PickupLastName() string { return v.LastName }

func (v *RiverConnection) PickupBio() string { return v.Bio }

func (v *RiverConnection) GetKey(key string) string {
	return v.Delegate.Get(key)
}

func (v *RiverConnection) SetKey(key, value string) {
	v.Delegate.Set(key, value)
}
