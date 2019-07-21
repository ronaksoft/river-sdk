package repo

/*
   Creation Time: 2019 - Jul - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type MessageSearch struct {
	Type   string `json:"type"`
	Body   string `json:"body"`
	PeerID int64  `json:"peer_id"`
}

type UserSearch struct {
	Type      string `json:"type"`
	FirstName string `json:"fn"`
	LastName  string `json:"ln"`
	Username  string `json:"un"`
	PeerID    int64  `json:"peer_id"`
}

type ContactSearch struct {
	Type      string `json:"type"`
	FirstName string `json:"fn"`
	LastName  string `json:"ln"`
	Username  string `json:"un"`
	Phone     string `json:"phone"`
}

type Group struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	PeerID int64  `json:"peer_id"`
}