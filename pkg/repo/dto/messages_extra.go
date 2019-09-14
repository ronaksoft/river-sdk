package dto

type MessagesExtra struct {
	dto
	PeerID   int64  `json:"PeerID"`
	PeerType int32  `json:"PeerType"`
	ScrollID int64  `json:"ScrollID"`
	Holes    []byte `json:"Holes"`
}