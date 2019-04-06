package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
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
func (s *ImportContact) Play(act shared.Acter) {
	if len(act.GetPeers()) > 0 {
		s.log(act, "Actor already have Peers", 0, 0)
		return
	}
	if act.GetAuthID() == 0 {
		s.AddJobs(1)
		success := Play(act, NewCreateAuthKey(false))
		if !success {

			s.failed(act, 0, 0, "Play() : failed at pre requested scenario CreateAuthKey")
			return
		}
		s.wait.Done()
	}
	if act.GetUserID() == 0 {
		s.AddJobs(1)
		success := Play(act, NewLogin(false))
		if !success {
			s.failed(act, 0, 0, "Play() : failed at pre requested scenario Login")
			return
		}
		s.wait.Done()
	}

	phoneList := act.GetPhoneList()
	s.AddJobs(len(phoneList))
	for _, p := range phoneList {
		act.ExecuteRequest(s.contactImport(p, act))
	}
}

// contactImport : Step 1
func (s *ImportContact) contactImport(phone string, act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
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

			if s.isFinal {
				err := act.Save()
				if err != nil {
					s.log(act, "contactImport() Actor.Save(), Err : "+err.Error(), elapsed, resp.RequestID)
				}
			}
			s.completed(act, elapsed, resp.RequestID, "contactImport() Success")
		} else {
			// TODO : Reporter failed
			s.failed(act, elapsed, resp.RequestID, "contactImport() SuccessCB response is not ContactsImported , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
