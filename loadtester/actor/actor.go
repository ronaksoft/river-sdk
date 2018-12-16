package actor

import "git.ronaksoftware.com/ronak/riversdk/domain"

type Actor struct {
	AuthID  int64
	AuthKey []byte //[256]byte
	UserID  int64
	Peers   []PeerInfo

	OnError   domain.ErrorHandler
	OnMessage domain.OnMessageHandler
	OnUpdate  domain.OnUpdateHandler
}
