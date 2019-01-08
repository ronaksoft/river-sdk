package filemanager

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Send the data payload is binary
func (fm *FileManager) Send(msgEnvelope *msg.MessageEnvelope, cluster *msg.Cluster) (*msg.MessageEnvelope, error) {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = fm.authID
	protoMessage.MessageKey = make([]byte, 32)
	if fm.authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		fm.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: 234242, // TODO:: ServerSalt ?
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(time.Now().Unix()<<32 | fm.messageSeq)
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(fm.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(fm.authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	b, err := protoMessage.Marshal()
	reqBuff := bytes.NewBuffer(b)
	if err != nil {
		return nil, err
	}
	// SEND data
	// set timeout
	client := http.DefaultClient
	client.Timeout = domain.DEFAULT_REQUEST_TIMEOUT
	httpResp, err := client.Post(cluster.Domain, "application/protobuf", reqBuff)
	if err != nil {
		return nil, err
	}
	resBuff, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	res := new(msg.MessageEnvelope)
	err = res.Unmarshal(resBuff)

	return res, err
}
