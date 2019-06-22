package scenario

import (
	"crypto/rand"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	ronak "git.ronaksoftware.com/ronak/toolbox"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// getAccessHash
// This function generate AccessHash for UserID1 to access UserID2
func getAccessHash(userID1, userID2 int64) uint64 {
	if userID1 == userID2 {
		return 0
	}
	return trimU52(ronak.CRC64([]byte(fmt.Sprintf("/%d/x2374x/%d/", userID1, userID2))))
}

func trimU52(num uint64) uint64 {
	if num > 4503599627370496 {
		return num >> 12
	}
	return num
}

func wrapEnvelope(ctr int64, data []byte) *msg.MessageEnvelope {
	env := new(msg.MessageEnvelope)
	env.Constructor = ctr
	env.Message = data
	env.RequestID = uint64(shared.GetSeqID())
	return env
}

func InitConnect() (envelop *msg.MessageEnvelope) {
	req := new(msg.InitConnect)
	// req.ClientNonce = uint64(domain.SequentialUniqueID())
	req.ClientNonce = uint64(shared.GetSeqID())
	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_InitConnect, data)

	return
}

func InitConnectTest() (envelop *msg.MessageEnvelope) {
	req := new(msg.InitConnectTest)
	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_InitConnectTest, data)

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

	envelop = wrapEnvelope(msg.C_InitCompleteAuth, data)

	return
}

func AccountResetAuthorizations() (envelop *msg.MessageEnvelope) {
	req := new(msg.AccountResetAuthorization)
	req.AuthID = 0

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_AccountResetAuthorization, data)
	return
}
func AuthSendCode(phone string) (envelop *msg.MessageEnvelope) {
	req := new(msg.AuthSendCode)
	req.Phone = phone

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_AuthSendCode, data)

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

	envelop = wrapEnvelope(msg.C_AuthRegister, data)

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

	envelop = wrapEnvelope(msg.C_AuthLogin, data)

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
	req.Body = "A" // strconv.FormatInt(req.RandomID, 10)
	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}
	envelop = wrapEnvelope(msg.C_MessagesSend, data)
	return
}

func GroupsCreate(title string, peers []*shared.PeerInfo) (envelop *msg.MessageEnvelope) {
	req := new(msg.GroupsCreate)
	req.Title = title
	for _, p := range peers {
		if p.PeerType != msg.PeerUser {
			continue
		}
		req.Users = append(req.Users, &msg.InputUser{
			UserID:     p.PeerID,
			AccessHash: p.AccessHash,
		})
	}
	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}
	envelop = wrapEnvelope(msg.C_GroupsCreate, data)
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

	envelop = wrapEnvelope(msg.C_ContactsImport, data)

	return
}

func AuthRecallReq() (envelop *msg.MessageEnvelope) {
	req := new(msg.AuthRecall)

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_AuthRecall, data)

	return
}

func GetPeerInfo(userID, peerID int64, peerType msg.PeerType) (peerInfo *shared.PeerInfo) {
	accessHash := uint64(0)
	switch peerType {
	case msg.PeerSelf:
		accessHash = 0
	case msg.PeerUser:
		accessHash = getAccessHash(userID, peerID)
	case msg.PeerGroup:
		accessHash = 0
	case msg.PeerSuperGroup:
		accessHash = 0
	case msg.PeerChannel:
		accessHash = 0
	}

	peerInfo = &shared.PeerInfo{
		AccessHash: accessHash,
		Name:       "ZZ",
		PeerID:     peerID,
		PeerType:   peerType,
	}

	return
}

func FileSavePart() (envelop *msg.MessageEnvelope, fileID int64, fileParts int32) {
	fileID = shared.GetSeqID()
	fileParts = 1
	req := new(msg.FileSavePart)
	req.FileID = fileID
	req.PartID = 1
	req.TotalParts = 1
	req.Bytes = make([]byte, domain.FilePayloadSize)
	rand.Read(req.Bytes)
	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_FileSavePart, data)

	return
}

func MessageSendMedia(fileID int64, totalParts int32, peer *shared.PeerInfo) (envelop *msg.MessageEnvelope) {
	req := new(msg.MessagesSendMedia)
	req.Peer = &msg.InputPeer{
		AccessHash: peer.AccessHash,
		ID:         peer.PeerID,
		Type:       peer.PeerType,
	}
	req.RandomID = shared.GetSeqID()

	req.MediaType = msg.InputMediaTypeUploadedDocument
	req.ReplyTo = 0
	req.ClearDraft = true
	attribFile := msg.DocumentAttributeFile{
		Filename: "N.raw",
	}
	attribFileBuff, _ := attribFile.Marshal()
	attrib := &msg.DocumentAttribute{
		Type: msg.AttributeTypeFile,
		Data: attribFileBuff,
	}
	media := msg.InputMediaUploadedDocument{
		Attributes: []*msg.DocumentAttribute{attrib},
		Caption:    "C",
		File:       &msg.InputFile{FileID: fileID, FileName: "N.raw", TotalParts: totalParts, MD5Checksum: ""},
		MimeType:   "application/raw",
		Stickers:   nil,
		Thumbnail:  nil,
	}
	mediaData, err := media.Marshal()
	if err != nil {
		panic(err)
	}
	req.MediaData = mediaData

	data, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	envelop = wrapEnvelope(msg.C_MessagesSendMedia, data)
	return
}
