package scenario

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

func wrapEnvelop(ctr int64, data []byte) *msg.MessageEnvelope {
	env := new(msg.MessageEnvelope)
	env.Constructor = ctr
	env.Message = data
	env.RequestID = uint64(domain.SequentialUniqueID())
	return env
}

func AuthSendCode(phone string) (envelop *msg.MessageEnvelope) {
	req := new(msg.AuthSendCode)
	req.Phone = phone

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_AuthSendCode, data)

	return
}

func AuthRegister(phone, code, hash string) (envelop *msg.MessageEnvelope) {
	req := new(msg.AuthRegister)
	req.Phone = phone
	req.PhoneCode = code
	req.PhoneCodeHash = hash
	req.FirstName = phone
	req.LastName = phone

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_AuthRegister, data)

	return
}
