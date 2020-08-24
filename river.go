package riversdk

import (
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	mon "git.ronaksoft.com/ronak/riversdk/internal/monitoring"
	fileCtrl "git.ronaksoft.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoft.com/ronak/riversdk/pkg/repo"
	"git.ronaksoft.com/ronak/riversdk/pkg/salt"
	"go.uber.org/zap"
	"sort"
	"strconv"
	"sync"
	"time"

	"git.ronaksoft.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoft.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoft.com/ronak/riversdk/pkg/ctrl_sync"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
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
	// DbID is used to save data for different accounts in separate databases. Could be used for multi-account cases.
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
	LogDirectory           string

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
// by this SDK is Badger V2. 'repo' is the package name selected to handle repository functions.
type River struct {
	// Connection Info
	ConnInfo *RiverConnection
	// Device Token
	DeviceToken *msg.AccountRegisterDevice

	// localCommands can be satisfied by client cache
	localCommands map[int64]domain.LocalMessageHandler
	// realTimeCommands should not passed to queue to send they should directly pass to networkController
	realTimeCommands map[int64]bool

	// Team
	teamID         int64
	teamAccessHash uint64

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

func (r *River) saveDeviceToken() {
	val, err := json.Marshal(r.DeviceToken)
	if err != nil {
		logs.Error("We got error on marshalling device token", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.SkDeviceToken, string(val))
	if err != nil {
		logs.Error("We got error on saving device token in db", zap.Error(err))
		return
	}
}

func (r *River) loadDeviceToken() {
	r.DeviceToken = new(msg.AccountRegisterDevice)
	str, err := repo.System.LoadString(domain.SkDeviceToken)
	if err != nil {
		logs.Info("We did not find device token")
		return
	}
	err = json.Unmarshal([]byte(str), r.DeviceToken)
	if err != nil {
		logs.Warn("We couldn't unmarshal device token", zap.Error(err))
	}
}

func (r *River) onNetworkConnect() (err error) {
	domain.WindowLog(fmt.Sprintf("Connected: %s", domain.StartTime.Format(time.Kitchen)))
	var serverUpdateID int64
	waitGroup := &sync.WaitGroup{}
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
		r.networkCtrl.SetAuthRecalled(true)
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
			waitGroup.Add(1)
			r.syncCtrl.GetContacts(waitGroup)
			waitGroup.Wait()
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
		// We don't do anything just log, but client must call logout
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

	err := domain.Try(3, time.Millisecond*500, func() error {
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
		msg.C_GroupsToggleAdmins:            r.groupToggleAdmin,
		msg.C_ContactsImport:                r.contactsImport,
		msg.C_ContactsAdd:                   r.contactsAdd,
		msg.C_ContactsDelete:                r.contactsDelete,
		msg.C_ContactsDeleteAll:             r.contactsDeleteAll,
		msg.C_ContactsGetTopPeers:           r.contactsGetTopPeers,
		msg.C_ContactsResetTopPeer:          r.contactsResetTopPeer,
		msg.C_MessagesReadContents:          r.messagesReadContents,
		msg.C_UsersGetFull:                  r.usersGetFull,
		msg.C_AccountRemovePhoto:            r.accountRemovePhoto,
		msg.C_GroupsRemovePhoto:             r.groupRemovePhoto,
		msg.C_MessagesSendMedia:             r.messagesSendMedia,
		msg.C_ClientSendMessageMedia:        r.clientSendMessageMedia,
		msg.C_MessagesSaveDraft:             r.messagesSaveDraft,
		msg.C_MessagesClearDraft:            r.messagesClearDraft,
		msg.C_LabelsGet:                     r.labelsGet,
		msg.C_LabelsListItems:               r.labelsListItems,
		msg.C_LabelsAddToMessage:            r.labelAddToMessage,
		msg.C_LabelsRemoveFromMessage:       r.labelRemoveFromMessage,
		msg.C_ClientGlobalSearch:            r.clientGlobalSearch,
		msg.C_ClientContactSearch:           r.clientContactSearch,
		msg.C_ClientGetCachedMedia:          r.clientGetCachedMedia,
		msg.C_ClientClearCachedMedia:        r.clientClearCachedMedia,
		msg.C_ClientGetMediaHistory:         r.clientGetMediaHistory,
		msg.C_ClientGetLastBotKeyboard:      r.clientGetLastBotKeyboard,
		msg.C_ClientGetRecentSearch:         r.clientGetRecentSearch,
		msg.C_ClientPutRecentSearch:         r.clientPutRecentSearch,
		msg.C_ClientRemoveAllRecentSearches: r.clientRemoveAllRecentSearches,
		msg.C_ClientRemoveRecentSearch:      r.clientRemoveRecentSearch,
		msg.C_GifGetSaved:                   r.gifGetSaved,
		msg.C_GifSave:                       r.gifSave,
		msg.C_GifDelete:                     r.gifDelete,
		msg.C_SystemGetConfig:               r.systemGetConfig,
	}
}

// PublicKey ...
// easyjson:json
type PublicKey struct {
	N           string
	FingerPrint int64
	E           uint32
}

// DHGroup ...
// easyjson:json
type DHGroup struct {
	Prime       string
	Gen         int32
	FingerPrint int64
}

// ServerKeys ...
// easyjson:json
type ServerKeys struct {
	PublicKeys []PublicKey
	DHGroups   []DHGroup
}

// GetPublicKey ...
func (v *ServerKeys) GetPublicKey(keyFP int64) (PublicKey, error) {
	logs.Info("Public Keys loaded",
		zap.Any("Public Keys", v.PublicKeys),
		zap.Int64("keyFP", keyFP),
	)

	for _, pk := range v.PublicKeys {
		if pk.FingerPrint == keyFP {
			return pk, nil
		}
	}
	return PublicKey{}, domain.ErrNotFound
}

// GetDhGroup ...
func (v *ServerKeys) GetDhGroup(keyFP int64) (DHGroup, error) {
	for _, dh := range v.DHGroups {
		if dh.FingerPrint == keyFP {
			return dh, nil
		}
	}
	return DHGroup{}, domain.ErrNotFound
}

// RiverConnection connection info
// easyjson:json
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

// save RiverConfig interface func
func (v *RiverConnection) Save() {
	logs.Debug("ConnInfo saved.")
	b, _ := v.MarshalJSON()
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
