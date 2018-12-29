package actor

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/executer"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
)

type Actor struct {
	AuthID  int64
	AuthKey []byte //[256]byte
	UserID  int64
	Peers   []PeerInfo

	netCtrl *controller.CtrlNetwork
	exec    *executer.Executer
}

// NewActor create new actor instance
func NewActor(userID int64, authID int64, authKey []byte) (*Actor, error) {

	act := &Actor{
		UserID:  userID,
		AuthID:  authID,
		AuthKey: authKey,
		Peers:   make([]PeerInfo, 0),
	}
	act.netCtrl = controller.NewCtrlNetwork(authID, authKey, act.onMessage, act.onUpdate, act.onError)
	err := act.netCtrl.Start()
	if err != nil {
		return act, err
	}
	act.exec = executer.NewExecuter(act.netCtrl)

	return act, nil
}

// onMessage check requestCallbacks and call callbacks
func (act *Actor) onMessage(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		req := act.exec.GetRequest(m.RequestID)
		if req != nil {
			select {
			case req.ResponseWaitChannel <- m:
				log.LOG_Debug("Actor::onMessage() callback singnal sent")
			default:
				log.LOG_Debug("Actor::onMessage() callback is skipped probably timedout before")
			}
			act.exec.RemoveRequest(m.RequestID)
		} else {
			log.LOG_Warn("Actor::onMessage() callback does not exists",
				zap.Uint64("RequestID", m.RequestID),
			)
		}
	}
}

func (act *Actor) onUpdate(updates []*msg.UpdateContainer) {
	for _, cnt := range updates {
		for _, u := range cnt.Updates {
			// TODO : Implement actors update reactions
			if u.Constructor == msg.C_UpdateNewMessage {
			}
		}
	}
}

func (act *Actor) onError(err *msg.Error) {
	// TODO : Add reporter error log
	log.LOG_Warn("Actor::onError() ", zap.String("Error", err.String()))
}
