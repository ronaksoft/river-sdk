package notification

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"git.ronaksoft.com/river/sdk/module"
	"github.com/ronaksoft/rony"
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

func (r *notification) clientDismissNotification(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientDismissNotification{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := repo.Notifications.SetNotificationDismissTime(domain.GetTeamID(in), req.Peer, req.Ts)
	if err != nil {
		r.Log().Error("got error on set client dismiss notification", zap.Error(err))
	}
}

func (r *notification) clientGetNotificationDismissTime(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientDismissNotification{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	ts, err := repo.Notifications.GetNotificationDismissTime(domain.GetTeamID(in), req.Peer)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	res := &msg.ClientNotificationDismissTime{
		Ts: ts,
	}
	out.Constructor = msg.C_ClientNotificationDismissTime
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(da.OnComplete, out)
}
