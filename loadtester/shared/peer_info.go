package shared

import "git.ronaksoftware.com/ronak/riversdk/msg"

type PeerInfo struct {
	PeerID     int64
	PeerType   msg.PeerType
	AccessHash uint64
	Name       string
}
