package bot

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/module"
	"github.com/ronaksoft/rony"
	"go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type bot struct {
	module.Base
}

func New() *bot {
	r := &bot{}
	r.RegisterMessageAppliers(
		map[int64]domain.MessageApplier{
			msg.C_BotResults: r.botResults,
		},
	)
	return r
}

func (r *bot) Name() string {
	return module.Bot
}
func (r *bot) botResults(e *rony.MessageEnvelope) {
	br := &msg.BotResults{}
	err := br.Unmarshal(e.Message)
	if err != nil {
		r.Log().Error("couldn't unmarshal BotResults", zap.Error(err))
		return
	}

	for _, m := range br.Results {
		if m == nil || m.Message == nil || m.Type != msg.MediaType_MediaTypeDocument || m.Message.MediaData == nil {
			r.Log().Info("botResults message or media is nil or not mediaDocument", zap.Error(err))
			continue
		}

		md := &msg.MediaDocument{}
		err := md.Unmarshal(m.Message.MediaData)
		if err != nil {
			r.Log().Error("couldn't unmarshal BotResults MediaDocument", zap.Error(err))
			continue
		}

		err = repo.Files.SaveMessageMediaDocument(md)

		if err != nil {
			r.Log().Error("couldn't save botResults media document", zap.Error(err))
		}
	}
}
