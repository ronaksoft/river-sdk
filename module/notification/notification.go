package notification

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/module"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
	"go.uber.org/zap"
)

/**
 * @created 25/05/2021 - 13:04
 * @project sdk
 * @author Reza Pilehvar
 */

type notification struct {
	module.Base
}

func New() *notification {
	r := &notification{}
	r.RegisterHandlers(
		map[int64]request.LocalHandler{
			msg.C_ClientDismissNotification:        r.clientDismissNotification,
			msg.C_ClientGetNotificationDismissTime: r.clientGetNotificationDismissTime,
		},
	)
	return r
}

func (r *notification) Name() string {
	return module.Notification
}

func (r *notification) clientDismissNotification(da request.Callback) {
	req := &msg.ClientDismissNotification{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := repo.Notifications.SetNotificationDismissTime(da.TeamID(), req.Peer, req.Ts)
	if err != nil {
		r.Log().Error("got error on set client dismiss notification", zap.Error(err))
	}
}

func (r *notification) clientGetNotificationDismissTime(da request.Callback) {
	req := &msg.ClientDismissNotification{}
	if err := da.RequestData(req); err != nil {
		return
	}

	ts, err := repo.Notifications.GetNotificationDismissTime(da.TeamID(), req.Peer)
	if err != nil {
		da.Response(rony.C_Error, errors.New("00", err.Error()))
		return
	}

	res := &msg.ClientNotificationDismissTime{
		Ts: ts,
	}
	da.Response(msg.C_ClientNotificationDismissTime, res)
}
