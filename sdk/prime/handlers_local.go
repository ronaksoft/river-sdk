package riversdk

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/ronaksoft/rony"
)

func (r *River) systemGetConfig(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	out.Fill(out.RequestID, msg.C_SystemConfig, domain.SysConfig)
	successCB(out)
}
