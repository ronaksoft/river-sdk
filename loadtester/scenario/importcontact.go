package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// ImportContact scenario
type ImportContact struct {
	Scenario
}

// NewImportContact create new instance
func NewImportContact() *ImportContact {
	s := new(ImportContact)
	return s
}

// Play execute ImportContact scenario
func (s *ImportContact) Play(act shared.Acter) {
	if act.GetAuthID() == 0 {
		Play(act, NewCreateAuthKey())
	}
	if act.GetUserID() == 0 {
		Play(act, NewLogin())
	}
	for _, p := range act.GetPhoneList() {
		s.wait.Add(1)
		act.ExecuteRequest(s.contactImport(p, act))
	}
}

// contactImport : Step 1
func (s *ImportContact) contactImport(phone string, act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := ContactsImport(phone)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
		s.failed("contactImport() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if s.isErrorResponse(resp) {
			return
		}
		if resp.Constructor == msg.C_ContactsImported {
			x := new(msg.ContactsImported)
			x.Unmarshal(resp.Message)
			// TODO : Complete scenario
			peers := make([]*shared.PeerInfo, 0, len(x.Users))
			for _, u := range x.Users {
				peers = append(peers, &shared.PeerInfo{
					PeerID:     u.ID,
					PeerType:   msg.PeerUser,
					AccessHash: u.AccessHash,
					Name:       u.FirstName + " " + u.LastName,
				})
			}
			act.SetPeers(peers)
			s.completed("contactImport() Success")
		} else {
			// TODO : Reporter failed
			s.failed("contactImport() SuccessCB response is not ContactsImported")
		}
	}

	return reqEnv, successCB, timeoutCB
}
