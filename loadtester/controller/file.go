package controller

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

var (
	messageSeq = int64(1)
)

// ExecuteFileRequest encrypt and send request to server and receive and decrypt its response
func ExecuteFileRequest(msgEnvelope *msg.MessageEnvelope, act shared.Acter) (*msg.MessageEnvelope, error) {
	authID, authKey := act.GetAuthInfo()
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = authID
	protoMessage.MessageKey = make([]byte, 32)
	if authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		msgSeq := atomic.AddInt64(&messageSeq, 1)
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: 234242, // TODO:: ServerSalt ?
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(time.Now().Unix()<<32 | msgSeq)
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(authKey, unencryptedBytes)
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
	client.Timeout = shared.DefaultTimeout

	// Send Data
	httpResp, err := client.Post(shared.DefaultFileServerURL, "application/protobuf", reqBuff)
	if err != nil {
		return nil, err
	}
	// Read response
	resBuff, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// metric
	shared.Metrics.Counter(shared.CntReceive).Add(float64(len(resBuff)))

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
	decryptedBytes, err := domain.Decrypt(authKey, res.MessageKey, res.Payload)

	receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
	err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return receivedEncryptedPayload.Envelope, nil
}
