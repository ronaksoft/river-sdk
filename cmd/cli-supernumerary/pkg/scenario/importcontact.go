package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// ImportContact scenario
type ImportContact struct {
	Scenario
}

// NewImportContact create new instance
func NewImportContact(isFinal bool) shared.Screenwriter {
	s := new(ImportContact)
	s.isFinal = isFinal
	return s
}

// Play execute ImportContact scenario
func (s *ImportContact) Play(act shared.Actor) {
	if len(act.GetPeers()) > 0 {
		s.log(act, "Actor already have Peers", 0, 0)
		return
	}
	if act.GetAuthID() == 0 {
		s.log(act, "AuthID is ZERO, Scenario Failed", 0, 0)
		return
	}
	if act.GetUserID() == 0 {
		s.log(act, "UserID is ZERO, Scenario Failed", 0, 0)
		return
	}

	phoneList := act.GetPhoneList()
	s.AddJobs(len(phoneList))
	for _, p := range phoneList {
		act.ExecuteRequest(s.contactImport(p, act))
	}
}

// contactImport : Step 1
func (s *ImportContact) contactImport(phone string, act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := ContactsImport(phone)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_ContactsImport, elapsed)
		s.failed(act, elapsed, requestID, "contactImport() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_ContactsImport, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "contactImport()") {
			return
		}
		if resp.Constructor == msg.C_ContactsImported {
			x := new(msg.ContactsImported)
			_ = x.Unmarshal(resp.Message)
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

			if s.isFinal {
				err := act.Save()
				if err != nil {
					s.log(act, "contactImport() Actor.save(), Err : "+err.Error(), elapsed, resp.RequestID)
				}
			}
			s.completed(act, elapsed, resp.RequestID, "contactImport() Success")
		} else {
			s.failed(act, elapsed, resp.RequestID, "contactImport() SuccessCB response is not ContactsImported , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
