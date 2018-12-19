package actor

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/executer"
	"git.ronaksoftware.com/ronak/riversdk/msg"
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

func (act *Actor) onMessage(messages []*msg.MessageEnvelope) {

}

func (act *Actor) onUpdate(messages []*msg.UpdateContainer) {

}

func (act *Actor) onError(u *msg.Error) {

}
