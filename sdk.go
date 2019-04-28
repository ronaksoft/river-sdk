package riversdk

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/filemanager"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"

	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_queue"
	"git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_sync"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"

	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"

	"github.com/monnand/dhkx"
	"go.uber.org/zap"
)

var (
	_ServerKeys ServerKeys
)

// SetConfig ...
// This function must be called before any other function, otherwise it panics
func (r *River) SetConfig(conf *RiverConfig) {
	r.lastOutOfSyncTime = time.Now().Add(1 * time.Second)
	r.chOutOfSyncUpdates = make(chan []*msg.UpdateContainer, 500)

	r.registerCommandHandlers()
	r.delegates = make(map[int64]RequestDelegate)

	// init delegates
	r.mainDelegate = conf.MainDelegate

	// set loglevel
	logs.SetLogLevel(conf.LogLevel)

	// check logger
	if conf.Logger != nil {
		// set logger
		r.logger = conf.Logger
		logs.SetHook(func(logLevel int, msg string) {
			r.logger.Log(logLevel, msg)
		})
	}

	// set log file path
	if conf.DocumentLogDirectory != "" {
		t := time.Now()
		fName := fmt.Sprintf("%d-%02d-%02d.log", t.Year(), t.Month(), t.Day())
		logDir := conf.DocumentLogDirectory
		// support IOS file path
		if strings.HasPrefix(logDir, "file://") {
			logDir = logDir[7:]
		}
		logFilePath := path.Join(logDir, fName)
		_ = logs.SetLogFilePath(logFilePath)
		logs.Info("SetConfig() ", zap.String("Log File Path", logFilePath))
	}

	// init UI Executor
	uiexec.InitUIExec()

	// Initialize Database
	_ = os.MkdirAll(conf.DbPath, os.ModePerm)
	conf.DbPath = strings.TrimRight(conf.DbPath, "/ ")

	// Initialize DB replaced with ORM
	var err error
	err = repo.InitRepo("sqlite3", fmt.Sprintf("%s/%s.db", conf.DbPath, conf.DbID))
	if err != nil {
		logs.Fatal("River::SetConfig() faild to initialize DB context",
			zap.String("Error", err.Error()),
		)
	}

	// init riverConfigs this should be after connect to DB
	r.loadSystemConfig()

	// load DeviceToken
	r.loadDeviceToken()

	// Initialize realtime requests
	r.realTimeCommands = map[int64]bool{
		msg.C_MessagesSetTyping: true,
		// msg.C_AuthRecall:        true,
		// msg.C_InitConnect:       true,
		// msg.C_InitCompleteAuth:  true,
	}

	// Initialize filemanager
	fileServerAddress := ""
	if strings.HasSuffix(conf.ServerEndpoint, "/") {
		fileServerAddress = conf.ServerEndpoint + "file"
	} else {
		fileServerAddress = conf.ServerEndpoint + "/file"
	}
	fileServerAddress = strings.Replace(fileServerAddress, "ws://", "http://", 1)
	filemanager.SetRootFolders(conf.DocumentAudioDirectory, conf.DocumentFileDirectory, conf.DocumentPhotoDirectory, conf.DocumentVideoDirectory, conf.DocumentCacheDirectory)

	filemanager.InitFileManager(fileServerAddress,
		r.onFileUploadCompleted,
		r.onFileProgressChanged,
		r.onFileDownloadCompleted,
		r.onFileUploadError,
		r.onFileDownloadError,
	)

	// Initialize Network Controller
	r.networkCtrl = network.NewController(
		network.Config{
			ServerEndpoint: conf.ServerEndpoint,
			PingTime:       time.Duration(conf.PingTimeSec) * time.Second,
			PongTimeout:    time.Duration(conf.PongTimeoutSec) * time.Second,
		},
	)
	r.networkCtrl.SetNetworkStatusChangedCallback(func(newQuality domain.NetworkStatus) {
		filemanager.Ctx().SetNetworkStatus(newQuality)
		if r.mainDelegate != nil {
			r.mainDelegate.OnNetworkStatusChanged(int(newQuality))
		}
	})

	// Initialize queueController
	if q, err := queue.NewController(r.networkCtrl, conf.QueuePath); err != nil {
		logs.Fatal("River::SetConfig() faild to initialize MessageQueue",
			zap.String("Error", err.Error()),
		)
	} else {
		r.queueCtrl = q
	}

	// Initialize Sync Controller
	r.syncCtrl = synchronizer.NewSyncController(
		synchronizer.Config{
			ConnInfo:    r.ConnInfo,
			NetworkCtrl: r.networkCtrl,
			QueueCtrl:   r.queueCtrl,
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
	if jsonBytes, err := ioutil.ReadFile(conf.ServerKeysFilePath); err != nil {
		logs.Fatal("River::SetConfig() faild to open server keys",
			zap.String("Error", err.Error()),
		)
	} else if err := _ServerKeys.UnmarshalJSON(jsonBytes); err != nil {
		logs.Fatal("River::SetConfig() faild to unmarshal server keys",
			zap.String("Error", err.Error()),
		)
	}

	// Initialize River Connection
	logs.Info("River::SetConfig() Load/Create New River Connection")

	if r.ConnInfo.UserID != 0 {
		r.syncCtrl.SetUserID(r.ConnInfo.UserID)
	}

	// Update Network Controller
	r.networkCtrl.SetErrorHandler(r.onGeneralError)
	r.networkCtrl.SetMessageHandler(r.onReceivedMessage)
	r.networkCtrl.SetUpdateHandler(r.onReceivedUpdate)
	r.networkCtrl.SetOnConnectCallback(r.onNetworkConnect)
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	// Update FileManager
	filemanager.Ctx().SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	filemanager.Ctx().LoadQueueFromDB()
}

func (r *River) Version() string {
	// TODO:: automatic generation
	return "0.8.1"
}

// Get deviceToken
func (r *River) loadDeviceToken() {
	r.DeviceToken = new(msg.AccountRegisterDevice)
	str, err := repo.Ctx().System.LoadString(domain.ColumnDeviceToken)
	if err != nil {
		logs.Error("River::loadDeviceToken() failed to fetch DeviceToken",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = json.Unmarshal([]byte(str), r.DeviceToken)
	if err != nil {
		logs.Error("River::loadDeviceToken() failed to unmarshal DeviceToken",
			zap.String("Error", err.Error()),
		)
	}
}

func (r *River) onNetworkConnect() {
	// Get Server Time and set server time difference
	timeReq := new(msg.SystemGetServerTime)
	timeReqBytes, _ := timeReq.Marshal()
	for {
		err := r.queueCtrl.ExecuteRealtimeCommand(
			uint64(domain.SequentialUniqueID()),
			msg.C_SystemGetServerTime,
			timeReqBytes,
			nil,
			r.onGetServerTime,
			true,
			false,
		)
		if err == nil {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

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
				r.onAuthRecalled,
				true,
				false,
			)
			if err == nil {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		if r.DeviceToken == nil || r.DeviceToken.Token == "" {
			logs.Warn("onNetworkConnect() Device Token is not set")
		}

		// import contact from server
		r.syncCtrl.ContactImportFromServer()
	}
}

func (r *River) onGeneralError(e *msg.Error) {
	// TODO:: calll external handler
	logs.Info("River::onGeneralError()",
		zap.String("Code", e.Code),
		zap.String("Item", e.Items),
	)

	if r.mainDelegate != nil {
		buff, _ := e.Marshal()
		r.mainDelegate.OnGeneralError(buff)
	}
}

// called when network flushes received messages
func (r *River) onReceivedMessage(msgs []*msg.MessageEnvelope) {
	// sort messages by requestID
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].RequestID < msgs[j].RequestID
	})

	// sync localDB with responses in the background
	go r.syncCtrl.MessageHandler(msgs)

	// check requestCallbacks and call callbacks
	for idx := range msgs {
		cb := domain.GetRequestCallback(msgs[idx].RequestID)
		if cb != nil {
			// if there was any listener maybe request already time-out
			logs.Debug("River::onReceivedMessage() Request callback found", zap.Uint64("RequestID", cb.RequestID))
			select {
			case cb.ResponseChannel <- msgs[idx]:
				logs.Debug("River::onReceivedMessage() passed to callback listener", zap.Uint64("RequestID", cb.RequestID))
			default:
				logs.Error("River::onReceivedMessage() there is no callback listener", zap.Uint64("RequestID", cb.RequestID))
			}
			domain.RemoveRequestCallback(msgs[idx].RequestID)
		} else {
			logs.Error("River::onReceivedMessage() callback does not exists",
				zap.Uint64("RequestID", msgs[idx].RequestID),
			)
		}
	}
}

// called when network flushes received updates
func (r *River) onReceivedUpdate(updateContainers []*msg.UpdateContainer) {
	// sort updateContainers
	sort.Slice(updateContainers, func(i, j int) bool {
		return updateContainers[i].MinUpdateID < updateContainers[j].MinUpdateID
	})

	for idx := range updateContainers {
		r.syncCtrl.UpdateHandler(updateContainers[idx])
	}
}

// onGetServerTime update client & server time difference
func (r *River) onGetServerTime(m *msg.MessageEnvelope) {
	if m.Constructor == msg.C_SystemServerTime {
		x := new(msg.SystemServerTime)
		err := x.Unmarshal(m.Message)
		if err != nil {
			logs.Error("onGetServerTime()", zap.Error(err))
			return
		}
		// TODO : get time difference and apply it later on send packets to server
		clientTime := time.Now().Unix()
		serverTime := x.Timestamp
		delta := serverTime - clientTime
		r.networkCtrl.SetClientTimeDifference(delta)
		logs.Debug("River::onGetServerTime()",
			zap.Int64("ServerTime", serverTime),
			zap.Int64("ClientTime", clientTime),
			zap.Int64("Difference", delta),
		)
	}
}

func (r *River) onFileProgressChanged(messageID, processedParts, totalParts int64, stateType domain.FileStateType) {
	percent := float64(processedParts) / float64(totalParts) * float64(100)

	logs.Debug("onFileProgressChanged()",
		zap.Int64("MsgID", messageID),
		zap.Float64("Percent", percent),
	)

	// Notify UI that upload is completed
	if stateType == domain.FileStateDownload {
		if r.mainDelegate != nil {
			r.mainDelegate.OnDownloadProgressChanged(messageID, processedParts, totalParts, percent)
		}
	} else if stateType == domain.FileStateUpload {
		if r.mainDelegate != nil {
			r.mainDelegate.OnUploadProgressChanged(messageID, processedParts, totalParts, percent)
		}
	}

}

func (r *River) onFileUploadCompleted(messageID, fileID, targetID int64,
	clusterID int32, totalParts int64,
	stateType domain.FileStateType,
	filePath string,
	req *msg.ClientSendMessageMedia,
	thumbFileID int64,
	thumbTotalParts int32,
) {
	logs.Debug("onFileUploadCompleted()",
		zap.Int64("messageID", messageID),
		zap.Int64("fileID", fileID),
	)
	// if total parts are greater than zero it means we actually uploaded new file
	// else the doc was already uploaded we called this just to notify ui that upload finished
	switch stateType {
	case domain.FileStateUpload:
		// Create SendMessageMedia Request
		x := new(msg.MessagesSendMedia)
		x.Peer = req.Peer
		x.ClearDraft = req.ClearDraft
		x.MediaType = req.MediaType
		x.RandomID = fileID
		x.ReplyTo = req.ReplyTo

		//
		switch x.MediaType {
		case msg.InputMediaTypeEmpty:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypeUploadedPhoto:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypePhoto:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypeContact:
			panic("SDK:onFileUploadCompleted() not implemented")
		case msg.InputMediaTypeUploadedDocument:

			doc := new(msg.InputMediaUploadedDocument)
			doc.MimeType = req.FileMIME
			doc.Attributes = req.Attributes

			doc.Caption = req.Caption
			doc.File = &msg.InputFile{
				FileID:      fileID,
				FileName:    req.FileName,
				MD5Checksum: "",
				TotalParts:  int32(totalParts),
			}

			if thumbFileID > 0 && thumbTotalParts > 0 {
				doc.Thumbnail = &msg.InputFile{
					FileID:      thumbFileID,
					FileName:    "thumb_" + req.FileName,
					MD5Checksum: "",
					TotalParts:  thumbTotalParts,
				}
			}

			x.MediaData, _ = doc.Marshal()

		case msg.InputMediaTypeDocument:
			panic("SDK:onFileUploadCompleted() not implemented")
		default:
			panic("SDK:onFileUploadCompleted() invalid input media type")
		}
		reqBuff, _ := x.Marshal()
		requestID := uint64(fileID)
		r.queueCtrl.ExecuteCommand(requestID, msg.C_MessagesSendMedia, reqBuff, nil, nil, false)

	case domain.FileStateUploadAccountPhoto:
		// TODO : AccountUploadPhoto
		x := new(msg.AccountUploadPhoto)
		x.File = &msg.InputFile{
			FileID:      fileID,
			FileName:    strconv.FormatInt(fileID, 10) + ".jpg",
			TotalParts:  int32(totalParts),
			MD5Checksum: "",
		}
		reqBuff, err := x.Marshal()
		if err != nil {
			logs.Error("SDK::onFileUploadCompleted() marshal AccountUploadPhoto", zap.Error(err))
			return
		}
		requestID := uint64(domain.SequentialUniqueID())
		successCB := func(m *msg.MessageEnvelope) {
			logs.Debug("AccountUploadPhoto success callback")
			if m.Constructor == msg.C_Bool {
				x := new(msg.Bool)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("AccountUploadPhoto success callback", zap.Error(err))
				}

			}
			if m.Constructor == msg.C_Error {
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err != nil {
					logs.Error("AccountUploadPhoto timeout callback", zap.Error(err))
				}
				logs.Error("AccountUploadPhoto timeout callback", zap.String("Code", x.Code), zap.String("Item", x.Items))
			}
		}
		timeoutCB := func() {
			logs.Debug("AccountUploadPhoto timeoput callback")
		}
		r.queueCtrl.ExecuteCommand(requestID, msg.C_AccountUploadPhoto, reqBuff, timeoutCB, successCB, false)
	case domain.FileStateUploadGroupPhoto:
		// TODO : GroupUploadPhoto
		x := new(msg.GroupsUploadPhoto)
		x.GroupID = targetID
		x.File = &msg.InputFile{
			FileID:      fileID,
			FileName:    strconv.FormatInt(fileID, 10) + ".jpg",
			TotalParts:  int32(totalParts),
			MD5Checksum: "",
		}
		reqBuff, err := x.Marshal()
		if err != nil {
			logs.Error("SDK::onFileUploadCompleted() marshal GroupUploadPhoto", zap.Error(err))
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

	// Notify UI that upload is completed
	if r.mainDelegate != nil {
		r.mainDelegate.OnUploadCompleted(messageID, filePath)
	}
}

func (r *River) onFileDownloadCompleted(messageID int64, filePath string, stateType domain.FileStateType) {
	logs.Info("onFileDownloadCompleted()", zap.Int64("MsgID", messageID), zap.String("FilePath", filePath))
	// update file path of documents that have same DocID
	go repo.Ctx().Files.UpdateFilePathByDocumentID(messageID, filePath)
	// Notify UI that download is completed
	if r.mainDelegate != nil {
		r.mainDelegate.OnDownloadCompleted(messageID, filePath)
	}
}

func (r *River) onFileUploadError(messageID, requestID int64, filePath string, err []byte) {
	x := new(msg.Error)
	x.Unmarshal(err)
	logs.Error("onFileUploadError() received Error response",
		zap.Int64("MsgID", messageID),
		zap.Int64("ReqID", requestID),
		zap.String("Code", x.Code),
		zap.String("Item", x.Items),
	)
	// Notify UI that upload encountered an error
	if r.mainDelegate != nil {
		r.mainDelegate.OnUploadError(messageID, requestID, filePath, err)
	}
}

func (r *River) onFileDownloadError(messageID, requestID int64, filePath string, err []byte) {
	x := new(msg.Error)
	x.Unmarshal(err)
	logs.Error("onFileDownloadError() received Error response",
		zap.Int64("MsgID", messageID),
		zap.Int64("ReqID", requestID),
		zap.String("Code", x.Code),
		zap.String("Item", x.Items),
	)
	// Notify UI that download encountered an error
	if r.mainDelegate != nil {
		r.mainDelegate.OnDownloadError(messageID, requestID, filePath, err)
	}
}

// onAuthRecalled update cluster info
func (r *River) onAuthRecalled(m *msg.MessageEnvelope) {
	if m.Constructor == msg.C_AuthRecalled {
		x := new(msg.AuthRecalled)
		err := x.Unmarshal(m.Message)
		if err != nil {
			logs.Error("onAuthRecalled()", zap.Error(err))
			return
		}
		// // TODO : get time difference and apply it later on send packets to server
		// clientTime := time.Now().Unix()
		// serverTime := x.Timestamp

	}
}

func (r *River) registerCommandHandlers() {
	r.localCommands = make(map[int64]domain.LocalMessageHandler)
	r.localCommands[msg.C_MessagesGetDialogs] = r.messagesGetDialogs
	r.localCommands[msg.C_MessagesGetDialog] = r.messagesGetDialog
	r.localCommands[msg.C_MessagesGetHistory] = r.messagesGetHistory
	r.localCommands[msg.C_MessagesSend] = r.messagesSend
	r.localCommands[msg.C_ClientSendMessageMedia] = r.clientSendMessageMedia
	r.localCommands[msg.C_ContactsGet] = r.contactsGet
	r.localCommands[msg.C_MessagesReadHistory] = r.messagesReadHistory
	r.localCommands[msg.C_UsersGet] = r.usersGet
	r.localCommands[msg.C_MessagesGet] = r.messagesGet
	r.localCommands[msg.C_AccountUpdateUsername] = r.accountUpdateUsername
	r.localCommands[msg.C_AccountUpdateProfile] = r.accountUpdateProfile
	r.localCommands[msg.C_AccountRegisterDevice] = r.accountRegisterDevice
	r.localCommands[msg.C_AccountUnregisterDevice] = r.accountUnregisterDevice
	r.localCommands[msg.C_AccountSetNotifySettings] = r.accountSetNotifySettings
	r.localCommands[msg.C_GroupsEditTitle] = r.groupsEditTitle
	r.localCommands[msg.C_MessagesClearHistory] = r.messagesClearHistory
	r.localCommands[msg.C_MessagesDelete] = r.messagesDelete
	r.localCommands[msg.C_GroupsAddUser] = r.groupAddUser
	r.localCommands[msg.C_GroupsDeleteUser] = r.groupDeleteUser
	r.localCommands[msg.C_GroupsGetFull] = r.groupsGetFull
	r.localCommands[msg.C_GroupsUpdateAdmin] = r.groupUpdateAdmin
	r.localCommands[msg.C_ContactsImport] = r.contactsImport
	r.localCommands[msg.C_MessagesReadContents] = r.messagesReadContents
	r.localCommands[msg.C_MessagesSendMedia] = r.messagesSendMedia
	r.localCommands[msg.C_UsersGetFull] = r.usersGetFull
	r.localCommands[msg.C_AccountRemovePhoto] = r.accountRemovePhoto
	r.localCommands[msg.C_GroupsRemovePhoto] = r.groupRemovePhoto

}

// Start ...
func (r *River) Start() error {
	// Start Controllers
	if err := r.networkCtrl.Start(); err != nil {
		logs.Error("River::Start()", zap.Error(err))
		return err
	}
	r.queueCtrl.Start()
	r.syncCtrl.Start()

	// Connect to Server
	go r.networkCtrl.Connect()

	return nil
}

// Stop ...
func (r *River) Stop() {
	// Disconnect from Server
	r.networkCtrl.Disconnect()

	// Stop Controllers
	r.syncCtrl.Stop()
	r.queueCtrl.Stop()
	r.networkCtrl.Stop()
	uiexec.Ctx().Stop()

	// Close database connection
	err := repo.Ctx().Close()
	logs.Debug("River::Stop() failed to close DB context",
		zap.String("Error", err.Error()),
	)
}

// ExecuteCommand ...
// This is a wrapper function to pass the request to the queueController, to be passed to networkController for final
// delivery to the server.
func (r *River) ExecuteCommand(constructor int64, commandBytes []byte, delegate RequestDelegate, blockingMode, serverForce bool) (requestID int64, err error) {
	if _, ok := msg.ConstructorNames[constructor]; !ok {
		return 0, domain.ErrInvalidConstructor
	}

	commandBytesDump := deepCopy(commandBytes)

	waitGroup := new(sync.WaitGroup)
	requestID = domain.SequentialUniqueID()
	logs.Debug("River::ExecuteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	// if function is in blocking mode set the waitGroup to block until the job is done, otherwise
	// save 'delegate' into delegates list to be fetched later.
	if blockingMode {
		waitGroup.Add(1)
		defer waitGroup.Wait()
	} else if delegate != nil {
		r.delegateMutex.Lock()
		r.delegates[requestID] = delegate
		r.delegateMutex.Unlock()
	}

	// Timeout Callback
	timeoutCallback := func() {
		if blockingMode {
			defer waitGroup.Done()
		}
		err = domain.ErrRequestTimeout
		delegate.OnTimeout(err)
		r.releaseDelegate(requestID)
	}

	// Success Callback
	successCallback := func(envelope *msg.MessageEnvelope) {
		if blockingMode {
			defer waitGroup.Done()
		}
		b, _ := envelope.Marshal()
		delegate.OnComplete(b)
		r.releaseDelegate(requestID)
	}

	// If this request must be sent to the server then executeRemoteCommand
	if serverForce {
		executeRemoteCommand(
			r,
			uint64(requestID),
			constructor,
			commandBytesDump,
			timeoutCallback,
			successCallback,
		)
		return
	}

	// If the constructor is a realtime command, then just send it to the server
	if _, ok := r.realTimeCommands[constructor]; ok {
		err = r.queueCtrl.ExecuteRealtimeCommand(
			uint64(requestID),
			constructor,
			commandBytesDump,
			timeoutCallback,
			successCallback,
			blockingMode,
			true,
		)
		if err != nil {
			logs.Error("ExecuteRealtimeCommand()", zap.Error(err))
			if delegate != nil {
				delegate.OnTimeout(err)
			}
		}
		return
	}

	// If the constructor is a local command then
	_, ok := r.localCommands[constructor]
	if ok {
		execBlock := func() {
			executeLocalCommand(
				r,
				uint64(requestID),
				constructor,
				commandBytesDump,
				timeoutCallback,
				successCallback,
			)
		}
		if blockingMode {
			execBlock()
		} else {
			go execBlock()
		}
		return
	}

	// If we reached here, then execute the remote commands
	executeRemoteCommand(
		r,
		uint64(requestID),
		constructor,
		commandBytesDump,
		timeoutCallback,
		successCallback,
	)

	return
}
func executeLocalCommand(r *River, requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	logs.Debug("River::executeLocalCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)

	in := new(msg.MessageEnvelope)
	out := new(msg.MessageEnvelope)
	in.Constructor = constructor
	in.Message = commandBytes
	in.RequestID = requestID
	out.RequestID = in.RequestID
	// double check
	if applier, ok := r.localCommands[constructor]; ok {
		applier(in, out, timeoutCB, successCB)
	}
}
func executeRemoteCommand(r *River, requestID uint64, constructor int64, commandBytes []byte, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	logs.Debug("River::executeRemoteCommand()",
		zap.String("Constructor", msg.ConstructorNames[constructor]),
	)
	r.queueCtrl.ExecuteCommand(requestID, constructor, commandBytes, timeoutCB, successCB, true)
}
func deepCopy(commandBytes []byte) []byte {
	// Takes a copy of commandBytes b4 IOS/Android GC/OS collect/alter them
	length := len(commandBytes)
	buff := make([]byte, length)
	copy(buff, commandBytes)
	return buff
}

func (r *River) releaseDelegate(requestID int64) {
	logs.Debug("River::releaseDelegate()",
		zap.Int64("RequestID", requestID),
	)
	r.delegateMutex.Lock()
	if _, ok := r.delegates[requestID]; ok {
		delete(r.delegates, requestID)
	}
	r.delegateMutex.Unlock()
}

// CreateAuthKey ...
// This function creates an AuthID and AuthKey to be used for transporting messages between client and server
func (r *River) CreateAuthKey() (err error) {
	logs.Debug("River::CreateAuthKey()")

	// Wait for network
	r.networkCtrl.WaitForNetwork()

	var clientNonce, serverNonce, serverPubFP, serverDHFP, serverPQ uint64
	// 1. Send InitConnect to Server
	req1 := new(msg.InitConnect)
	req1.ClientNonce = uint64(domain.SequentialUniqueID())
	req1Bytes, _ := req1.Marshal()
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)

	logs.Info("River::CreateAuthKey() 1st Step Started :: InitConnect")

	executeRemoteCommand(
		r,
		uint64(domain.SequentialUniqueID()),
		msg.C_InitConnect,
		req1Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *msg.MessageEnvelope) {
			defer waitGroup.Done()
			logs.Debug("River::CreateAuthKey() Success Callback Called")
			switch res.Constructor {
			case msg.C_InitResponse:
				x := new(msg.InitResponse)
				err = x.Unmarshal(res.Message)
				if err != nil {
					logs.Error("River::CreateAuthKey() Success Callback", zap.Error(err))
				}
				clientNonce = x.ClientNonce
				serverNonce = x.ServerNonce
				serverPubFP = x.RSAPubKeyFingerPrint
				serverDHFP = x.DHGroupFingerPrint
				serverPQ = x.PQ
				logs.Debug("River::CreateAuthKey() InitResponse Received",
					zap.Uint64("ServerNonce", serverNonce),
					zap.Uint64("ClientNounce", clientNonce),
					zap.Uint64("ServerDhFingerPrint", serverDHFP),
					zap.Uint64("ServerFingerPrint", serverPubFP),
				)
			case msg.C_Error:
				err = domain.ParseServerError(res.Message)
			default:
				err = domain.ErrInvalidConstructor
			}
		},
	)

	// Wait for 1st step to complete
	waitGroup.Wait()
	if err != nil {
		logs.Error("River::CreateAuthKey() InitConnect", zap.Error(err))
		return
	}
	logs.Info("River::CreateAuthKey() 1st Step Finished")

	// 2. Send InitCompleteAuth
	req2 := new(msg.InitCompleteAuth)
	req2.ServerNonce = serverNonce
	req2.ClientNonce = clientNonce

	// Generate DH Pub Key
	dhGroup, err := _ServerKeys.GetDhGroup(int64(serverDHFP))
	if err != nil {
		return err
	}
	dhPrime := big.NewInt(0)
	dhPrime.SetString(dhGroup.Prime, 16)

	dh := dhkx.CreateGroup(dhPrime, big.NewInt(int64(dhGroup.Gen)))
	clientDhKey, _ := dh.GeneratePrivateKey(rand.Reader)
	req2.ClientDHPubKey = clientDhKey.Bytes()

	p, q := domain.SplitPQ(big.NewInt(int64(serverPQ)))
	if p.Cmp(q) < 0 {
		req2.P = p.Uint64()
		req2.Q = q.Uint64()
	} else {
		req2.P = q.Uint64()
		req2.Q = p.Uint64()
	}
	logs.Debug("River::CreateAuthKey() PQ Split",
		zap.Uint64("P", req2.P),
		zap.Uint64("Q", req2.Q),
	)

	q2Internal := new(msg.InitCompleteAuthInternal)
	q2Internal.SecretNonce = []byte(domain.RandomID(16))

	serverPubKey, err := _ServerKeys.GetPublicKey(int64(serverPubFP))
	if err != nil {
		return err
	}
	n := big.NewInt(0)
	n.SetString(serverPubKey.N, 10)
	rsaPublicKey := rsa.PublicKey{
		N: n,
		E: int(serverPubKey.E),
	}
	decrypted, _ := q2Internal.Marshal()
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPublicKey, decrypted)
	if err != nil {
		logs.Error("River::CreateAuthKey() -> EncryptPKCS1v15()", zap.Error(err))
	}
	req2.EncryptedPayload = encrypted
	req2Bytes, _ := req2.Marshal()

	waitGroup.Add(1)
	logs.Info("River::CreateAuthKey() 2nd Step Started :: InitConnect")
	executeRemoteCommand(
		r,
		// r.executeRealtimeCommand(
		uint64(domain.SequentialUniqueID()),
		msg.C_InitCompleteAuth,
		req2Bytes,
		func() {
			defer waitGroup.Done()
			err = domain.ErrRequestTimeout
		},
		func(res *msg.MessageEnvelope) {
			defer waitGroup.Done()
			switch res.Constructor {
			case msg.C_InitAuthCompleted:
				x := new(msg.InitAuthCompleted)
				x.Unmarshal(res.Message)
				switch x.Status {
				case msg.InitAuthCompleted_OK:
					serverDhKey, err := dh.ComputeKey(dhkx.NewPublicKey(x.ServerDHPubKey), clientDhKey)
					if err != nil {
						logs.Error("River::CreateAuthKey() -> ComputeKey()", zap.Error(err))
						return
					}
					// r.ConnInfo.AuthKey = serverDhKey.Bytes()
					copy(r.ConnInfo.AuthKey[:], serverDhKey.Bytes())
					authKeyHash, _ := domain.Sha256(r.ConnInfo.AuthKey[:])
					r.ConnInfo.AuthID = int64(binary.LittleEndian.Uint64(authKeyHash[24:32]))

					var secret []byte
					secret = append(secret, q2Internal.SecretNonce...)
					secret = append(secret, byte(msg.InitAuthCompleted_OK))
					secret = append(secret, authKeyHash[:8]...)
					secretHash, _ := domain.Sha256(secret)

					if x.SecretHash != binary.LittleEndian.Uint64(secretHash[24:32]) {
						fmt.Println(x.SecretHash, binary.LittleEndian.Uint64(secretHash[24:32]))
						err = domain.ErrSecretNonceMismatch
						return
					}
				case msg.InitAuthCompleted_RETRY:
					// TODO:: Retry with new DHKey
				case msg.InitAuthCompleted_FAIL:
					err = domain.ErrAuthFailed
					return
				}
				r.ConnInfo.Save()
				r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
				filemanager.Ctx().SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
			case msg.C_Error:
				err = domain.ParseServerError(res.Message)
				return
			default:
				err = domain.ErrInvalidConstructor
				return
			}
		},
	)

	// Wait for 2nd step to complete
	waitGroup.Wait()

	// double set AuthID
	r.networkCtrl.SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])
	filemanager.Ctx().SetAuthorization(r.ConnInfo.AuthID, r.ConnInfo.AuthKey[:])

	return
}

// AddRealTimeRequest ...
func (r *River) AddRealTimeRequest(constructor int64) {
	r.realTimeCommands[constructor] = true
}

// RemoveRealTimeRequest ...
func (r *River) RemoveRealTimeRequest(constructor int64) {
	delete(r.realTimeCommands, constructor)
}

