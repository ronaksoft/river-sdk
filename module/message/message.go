package message

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/module"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type message struct {
	module.Base
}

func New() *message {
	r := &message{}
	r.RegisterHandlers(
		map[int64]domain.LocalHandler{
			msg.C_MessagesClearDraft:         r.messagesClearDraft,
			msg.C_MessagesClearHistory:       r.messagesClearHistory,
			msg.C_MessagesDelete:             r.messagesDelete,
			msg.C_MessagesDeleteReaction:     r.messagesDeleteReaction,
			msg.C_MessagesGet:                r.messagesGet,
			msg.C_MessagesGetDialog:          r.messagesGetDialog,
			msg.C_MessagesGetDialogs:         r.messagesGetDialogs,
			msg.C_MessagesGetHistory:         r.messagesGetHistory,
			msg.C_MessagesGetMediaHistory:    r.messagesGetMediaHistory,
			msg.C_MessagesReadContents:       r.messagesReadContents,
			msg.C_MessagesReadHistory:        r.messagesReadHistory,
			msg.C_MessagesSaveDraft:          r.messagesSaveDraft,
			msg.C_MessagesSend:               r.messagesSend,
			msg.C_MessagesSendMedia:          r.messagesSendMedia,
			msg.C_MessagesSendReaction:       r.messagesSendReaction,
			msg.C_MessagesToggleDialogPin:    r.messagesToggleDialogPin,
			msg.C_MessagesTogglePin:          r.messagesTogglePin,
			msg.C_ClientGetFrequentReactions: r.clientGetFrequentReactions,
			msg.C_ClientGetMediaHistory:      r.clientGetMediaHistory,
			msg.C_ClientSendMessageMedia:     r.clientSendMessageMedia,
			msg.C_ClientClearCachedMedia:     r.clientClearCachedMedia,
			msg.C_ClientGetCachedMedia:       r.clientGetCachedMedia,
			msg.C_ClientGetLastBotKeyboard:   r.clientGetLastBotKeyboard,
		},
	)
	r.RegisterUpdateAppliers(
		map[int64]domain.UpdateApplier{
			msg.C_UpdateDialogPinned:         r.updateDialogPinned,
			msg.C_UpdateDraftMessage:         r.updateDraftMessage,
			msg.C_UpdateDraftMessageCleared:  r.updateDraftMessageCleared,
			msg.C_UpdateMessageEdited:        r.updateMessageEdited,
			msg.C_UpdateMessageID:            r.updateMessageID,
			msg.C_UpdateMessagePinned:        r.updateMessagePinned,
			msg.C_UpdateMessagesDeleted:      r.updateMessagesDeleted,
			msg.C_UpdateNewMessage:           r.updateNewMessage,
			msg.C_UpdateNotifySettings:       r.updateNotifySettings,
			msg.C_UpdateReaction:             r.updateReaction,
			msg.C_UpdateReadHistoryInbox:     r.updateReadHistoryInbox,
			msg.C_UpdateReadHistoryOutbox:    r.updateReadHistoryOutbox,
			msg.C_UpdateReadMessagesContents: r.updateReadMessagesContents,
			msg.C_UpdatePhoneCallStarted:     r.updatePhoneCallStarted,
			msg.C_UpdatePhoneCallEnded:       r.updatePhoneCallEnded,
		},
	)
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_MessagesDialogs:      r.messagesDialogs,
			msg.C_MessagesMany:         r.messagesMany,
			msg.C_MessagesReactionList: r.reactionList,
		},
	)
	return r
}

func (r *message) Name() string {
	return module.Message
}
