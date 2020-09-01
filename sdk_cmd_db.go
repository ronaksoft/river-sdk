package riversdk

import (
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"git.ronaksoft.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

func (r *River) IsMessageExist(messageID int64) bool {
	message, _ := repo.Messages.Get(messageID)

	return message != nil
}

func (r *River) IsGifSaved(fileID int64, clusterID int32) bool {
	return repo.Gifs.IsSaved(clusterID, fileID)
}

func (r *River) GetRealTopMessageID(peerID int64, peerType int32) int64 {
	topMsgID, err := repo.Messages.GetTopMessageID(domain.GetCurrTeamID(), peerID, peerType)
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

func (r *River) ResetCalculatedImportHash() {
	err := repo.System.SaveInt(domain.SkContactsImportHash, 0)
	logs.ErrorOnErr("ResetCalculatedImportHash", err)
}
