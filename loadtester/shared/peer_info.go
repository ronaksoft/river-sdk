package shared

import "git.ronaksoftware.com/ronak/riversdk/msg"

// PeerInfo actor contact info
type PeerInfo struct {
	PeerID     int64
	PeerType   msg.PeerType
	AccessHash uint64
	Name       string
}
