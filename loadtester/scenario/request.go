package scenario

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

func wrapEnvelop(ctr int64, data []byte) *msg.MessageEnvelope {
	env := new(msg.MessageEnvelope)
	env.Constructor = ctr
	env.Message = data
	// env.RequestID = uint64(domain.SequentialUniqueID())
	env.RequestID = uint64(shared.GetSeqID())
	return env
}

func InitConnect() (envelop *msg.MessageEnvelope) {
	req := new(msg.InitConnect)
	//req.ClientNonce = uint64(domain.SequentialUniqueID())
	req.ClientNonce = uint64(shared.GetSeqID())
	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_InitConnect, data)

	return
}

func InitCompleteAuth(clientNonce, serverNonce, p, q uint64, dhPubKey, encPayload []byte) (envelop *msg.MessageEnvelope) {
	req := new(msg.InitCompleteAuth)

	req.ClientNonce = clientNonce
	req.ServerNonce = serverNonce
	req.P = p
	req.Q = q
	req.ClientDHPubKey = dhPubKey
	req.EncryptedPayload = encPayload

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_InitCompleteAuth, data)

	return
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

func AuthLogin(phone, code, hash string) (envelop *msg.MessageEnvelope) {
	req := new(msg.AuthLogin)
	req.Phone = phone
	req.PhoneCode = code
	req.PhoneCodeHash = hash

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_AuthLogin, data)

	return
}

func MessageSend(peer *shared.PeerInfo) (envelop *msg.MessageEnvelope) {
	req := new(msg.MessagesSend)
	req.Peer = &msg.InputPeer{
		AccessHash: peer.AccessHash,
		ID:         peer.PeerID,
		Type:       peer.PeerType,
	}
	// req.RandomID = domain.SequentialUniqueID()
	req.RandomID = shared.GetSeqID()
	req.Body = "A" //strconv.FormatInt(req.RandomID, 10)

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_MessagesSend, data)

	return
}

func ContactsImport(phone string) (envelop *msg.MessageEnvelope) {
	req := new(msg.ContactsImport)
	req.Contacts = []*msg.PhoneContact{
		&msg.PhoneContact{
			ClientID:  shared.GetSeqID(),
			FirstName: phone,
			LastName:  phone,
			Phone:     phone,
		},
	}
	req.Replace = true

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelop(msg.C_ContactsImport, data)

	return
}
