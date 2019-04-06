package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type SendFile struct {
	Scenario
}

// NewSendFile create new instance
func NewSendFile(isFinal bool) shared.Screenwriter {
	s := new(SendFile)
	s.isFinal = isFinal
	return s
}

// Play execute SendFile scenario
func (s *SendFile) Play(act shared.Acter) {
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
	if len(act.GetPeers()) == 0 {
		s.AddJobs(1)
		success := Play(act, NewImportContact(false))
		if !success {
			s.failed(act, 0, 0, "Play() : failed at pre requested scenario ImportContact")
			return
		}
		s.wait.Done()
	}
	peers := act.GetPeers()
	s.AddJobs(len(peers))
	for _, p := range peers {
		act.ExecuteRequest(s.fileUpload(act, p))
	}
}

// fileUpload : Step 1
func (s *SendFile) fileUpload(act shared.Acter, peer *shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {

	sw := time.Now()
	req, fileID, totalParts := FileSavePart()

	// upload file
	res, err := act.ExecFileRequest(req)
	if err != nil {
		s.failed(act, time.Since(sw), uint64(fileID), "ExecFileRequest failed "+err.Error())
		return nil, nil, nil
	}

	switch res.Constructor {
	case msg.C_Error:
		x := new(msg.Error)
		x.Unmarshal(res.Message)
		s.failed(act, time.Since(sw), uint64(fileID), "ExecFileRequest received C_Error { Code :"+x.Code+", Item :"+x.Items+"}")
	case msg.C_Bool:
		x := new(msg.Bool)
		x.Unmarshal(res.Message)
		if x.Result {

			reqEnv := MessageSendMedia(fileID, totalParts, peer)

			timeoutCB := func(requestID uint64, elapsed time.Duration) {
				// Reporter failed
				act.SetTimeout(msg.C_MessagesSendMedia, elapsed)
				s.failed(act, elapsed, requestID, "fileUpload() Timeout")
			}

			successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
				act.SetSuccess(msg.C_MessagesSendMedia, elapsed)
				if s.isErrorResponse(act, elapsed, resp, "fileUpload()") {
					return
				}
				if resp.Constructor == msg.C_MessagesSent {
					x := new(msg.MessagesSent)
					x.Unmarshal(resp.Message)

					// TODO : Complete Scenario
					s.completed(act, elapsed, resp.RequestID, "fileUpload() Success")
				} else {
					// TODO : Reporter failed
					s.failed(act, elapsed, resp.RequestID, "fileUpload() SuccessCB response is not MessagesSent, Constructor :"+msg.ConstructorNames[resp.Constructor])
				}
			}

			return reqEnv, successCB, timeoutCB
		}
	default:
		s.failed(act, time.Since(sw), uint64(fileID), "ExecFileRequest received unknown response constructor :"+msg.ConstructorNames[res.Constructor])
	}
	return nil, nil, nil
}
