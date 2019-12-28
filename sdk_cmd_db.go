package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

func (r *River) IsMessageExist(messageID int64) bool {
	message, _ := repo.Messages.Get(messageID)

	return message != nil
}

func (r *River) GetRealTopMessageID(peerID int64, peerType int32) int64 {
	topMsgID, err := repo.Messages.GetTopMessageID(peerID, peerType)
	if err != nil {
		logs.Error("SDK::GetRealTopMessageID() => Messages.GetTopMessageID()", zap.Error(err))
		return -1
	}
	return topMsgID
}

func (r *River) GetPinnedDialogsCount() int32 {
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
