package fileCtrl

import (
	"bytes"
	"context"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	"io/ioutil"
	"net/http"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

func (ctrl *Controller) SendWithContext(ctx context.Context, in *msg.MessageEnvelope) (*msg.MessageEnvelope, error) {
	// waitGroup := sync.WaitGroup{}
	// waitGroup.Add(1)
	// select {
	// case ctx.Done():
	//
	//
	// }
	panic("not implemented")
}

// Send encrypt and send request to server and receive and decrypt its response
func (ctrl *Controller) Send(msgEnvelope *msg.MessageEnvelope) (*msg.MessageEnvelope, error) {
	protoMessage := new(msg.ProtoMessage)
	protoMessage.AuthID = ctrl.authID
	protoMessage.MessageKey = make([]byte, 32)
	if ctrl.authID == 0 {
		protoMessage.Payload, _ = msgEnvelope.Marshal()
	} else {
		ctrl.messageSeq++
		encryptedPayload := msg.ProtoEncryptedPayload{
			ServerSalt: salt.Get(),
			Envelope:   msgEnvelope,
		}
		encryptedPayload.MessageID = uint64(time.Now().Unix()<<32 | ctrl.messageSeq)
		unencryptedBytes, _ := encryptedPayload.Marshal()
		encryptedPayloadBytes, _ := domain.Encrypt(ctrl.authKey, unencryptedBytes)
		messageKey := domain.GenerateMessageKey(ctrl.authKey, unencryptedBytes)
		copy(protoMessage.MessageKey, messageKey)
		protoMessage.Payload = encryptedPayloadBytes
	}

	b, err := protoMessage.Marshal()

	reqBuff := bytes.NewBuffer(b)
	if err != nil {
		return nil, err
	}

	// Set timeout
	client := &http.Client{}
	client.Timeout = domain.WebsocketRequestTime

	// Send Data
	httpResp, err := client.Post(ctrl.ServerEndpoint, "application/protobuf", reqBuff)
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
		return nil, err
	}
	if res.AuthID == 0 {
		receivedEnvelope := new(msg.MessageEnvelope)
		err = receivedEnvelope.Unmarshal(res.Payload)
		return receivedEnvelope, err
	}
	decryptedBytes, err := domain.Decrypt(ctrl.authKey, res.MessageKey, res.Payload)

	receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
	err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return receivedEncryptedPayload.Envelope, nil
}
