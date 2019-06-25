package riversdk

import "git.ronaksoftware.com/ronak/riversdk/pkg/repo"

func (r *River) IsMessageExist(messageID int64) bool {
	message := repo.Messages.GetMessage(messageID)

	return message != nil
}
