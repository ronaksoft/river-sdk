package filemanager

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

var (
	_MessageSeq int64
)

// Send the data payload is binary
func Send(msgEnvelope *msg.MessageEnvelope, cluster *msg.Cluster, authID int64, authKey []byte, chResult chan *msg.MessageEnvelope) error {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = authID
	protoMessage.MessageKey = make([]byte, 32)
	if authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		_MessageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: 234242, // TODO:: ServerSalt ?
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(time.Now().Unix()<<32 | _MessageSeq)
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	b, err := protoMessage.Marshal()
	reqBuff := bytes.NewBuffer(b)
	if err != nil {
		return err
	}
	// SEND data
	httpResp, err := http.Post(cluster.Domain, "application/protobuf", reqBuff)
	if err != nil {
		return err
	}
	resBuff, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	res := new(msg.MessageEnvelope)
	err = res.Unmarshal(resBuff)
	if err != nil {
		return err
	}
	select {
	case chResult <- res:
	default:
		log.LOG_Warn("filemanager::Send() no one listening at chResult channel")
	}
	return nil
}
