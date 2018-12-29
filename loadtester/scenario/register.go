package scenario

import (
	"strings"

	"git.ronaksoftware.com/ronak/riversdk/domain"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Actor receives scenario and execute its step one by one

type Register struct {
}

func (s *Register) NewRegister() {

}

func (s *Register) Execute() {

}

//sendCode : Step 1
func (s *Register) sendCode(phone string) (*msg.MessageEnvelope, domain.MessageHandler, domain.TimeoutCallback) {
	envReq := AuthSendCode(phone)
	timeoutCB := func() {
		// TODO : Reporter failed
	}
	successCB := func(resp *msg.MessageEnvelope) {
		// TODO : chain next request here
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			x.Unmarshal(resp.Message)
			s.register(x)
		} else {
			// TODO : Reporter failed
		}
	}

	return envReq, successCB, timeoutCB
}

// register : Step 2
func (s *Register) register(resp *msg.AuthSentCode) (*msg.MessageEnvelope, domain.MessageHandler, domain.TimeoutCallback) {
	if strings.HasSuffix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]
		envReq := AuthRegister(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func() {
			// TODO : Reporter failed
		}
		successCB := func(resp *msg.MessageEnvelope) {
			// TODO : chain next request here
			if resp.Constructor == msg.C_Bool {
				x := new(msg.Bool)
				x.Unmarshal(resp.Message)

				if x.Result {
					// TODO : reporter success
				} else {
					// TODO : Reporter failed
				}
			} else {
				// TODO : Reporter failed
			}
		}

		// TODO : Fix This
		return envReq, successCB, timeoutCB
	}

	// TODO : Reporter failed

	return nil, nil, nil
}
