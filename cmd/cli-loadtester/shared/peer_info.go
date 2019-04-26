package shared

import msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"

// PeerInfo actor contact info
type PeerInfo struct {
	PeerID     int64        `json:"peer_id"`
	PeerType   msg.PeerType `json:"peer_type"`
	AccessHash uint64       `json:"access_hash"`
	Name       string       `json:"name"`
}
