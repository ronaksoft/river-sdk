package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"time"
)

func (r *River) IsMessageExist(messageID int64) bool {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("IsMessageExists", time.Now().Sub(startTime))
	}()
	message := repo.Messages.Get(messageID)

	return message != nil
}

// GetRealTopMessageID returns max message id
func (r *River) GetRealTopMessageID(peerID int64, peerType int32) int64 {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetRealTopMessageID", time.Now().Sub(startTime))
	}()
	topMsgID, err := repo.Messages.GetTopMessageID(peerID, peerType)
	if err != nil {
		logs.Error("SDK::GetRealTopMessageID() => Messages.GetTopMessageID()", zap.Error(err))
		return -1
	}
	return topMsgID
}

func (r *River) GetPinnedDialogsCount() int32 {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetPinnedDialogCount", time.Now().Sub(startTime))
	}()
	dialogs := repo.Dialogs.GetPinnedDialogs()

	if dialogs != nil {
		return int32(len(dialogs))
	}

	return 0
}

// Run the garbage collector on db, this releases more space on the space, make sure to call
// this function when app is not busy
func (r *River) GC() {
	repo.GC()
}