package message

import (
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	queueCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	syncCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_sync"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/module"
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
	queueCtrl   *queueCtrl.Controller
	networkCtrl *networkCtrl.Controller
	fileCtrl    *fileCtrl.Controller
	syncCtrl    *syncCtrl.Controller
	sdk         module.SDK
}

func New() *message {
	return &message{}
}

func (r *message) Init(sdk module.SDK) {
	r.sdk = sdk
	r.networkCtrl = sdk.NetCtrl()
	r.queueCtrl = sdk.QueueCtrl()
	r.syncCtrl = sdk.SyncCtrl()
	r.fileCtrl = sdk.FileCtrl()

}

func (r *message) LocalHandlers() map[int64]domain.LocalMessageHandler {
	return map[int64]domain.LocalMessageHandler{
		msg.C_MessagesClearDraft:         r.messagesClearDraft,
		msg.C_MessagesClearHistory:       r.messagesClearHistory,
		msg.C_MessagesDelete:             r.messagesDelete,
		msg.C_MessagesDeleteReaction:     r.messagesDeleteReaction,
		msg.C_MessagesGet:                r.messagesGet,
		msg.C_MessagesGetDialog:          r.messagesGetDialog,
		msg.C_MessagesGetDialogs:         r.messagesGetDialogs,
		msg.C_MessagesGetHistory:         r.messagesGetHistory,
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
	}
}
