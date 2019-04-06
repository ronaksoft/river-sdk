package filemanager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

// Send encrypt and send request to server and receive and decrypt its response
func (fm *FileManager) Send(msgEnvelope *msg.MessageEnvelope) (*msg.MessageEnvelope, error) {
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

	//ioutil.WriteFile("dump.raw", b, os.ModePerm)

	reqBuff := bytes.NewBuffer(b)
	if err != nil {
		return nil, err
	}

	// Set timeout
	client := http.DefaultClient
	client.Timeout = domain.WebsocketRequestTime

	// Send Data
	httpResp, err := client.Post(fm.ServerEndpoint, "application/protobuf", reqBuff)
	if err != nil {
		return nil, err
	}
	// Read response
	resBuff, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// Decrypt response
	res := new(msg.ProtoMessage)
	err = res.Unmarshal(resBuff)
	if err != nil {
		return nil, fmt.Errorf("Error : %s , Response Dump : %s", err.Error(), string(resBuff))
	}
	if res.AuthID == 0 {
		receivedEnvelope := new(msg.MessageEnvelope)
		err = receivedEnvelope.Unmarshal(res.Payload)
		return receivedEnvelope, err
	}
	decryptedBytes, err := domain.Decrypt(fm.authKey, res.MessageKey, res.Payload)

	receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
	err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return receivedEncryptedPayload.Envelope, nil
}