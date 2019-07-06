package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
)

func (r *River) IsMessageExist(messageID int64) bool {
	message := repo.Messages.GetMessage(messageID)

	return message != nil
}

// GetRealTopMessageID returns max message id
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