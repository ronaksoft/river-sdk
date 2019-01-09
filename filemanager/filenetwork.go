package filemanager

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Send to file server cluster
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

	//ioutil.WriteFile("dump.raw", b, os.ModePerm)

	reqBuff := bytes.NewBuffer(b)
	if err != nil {
		return nil, err
	}

	// Set timeout
	client := http.DefaultClient
	client.Timeout = domain.DEFAULT_REQUEST_TIMEOUT
	if !strings.HasPrefix(cluster.Domain, "http") {
		cluster.Domain = "http://" + cluster.Domain
	}
	// Send Data
	httpResp, err := client.Post(cluster.Domain, "application/protobuf", reqBuff)
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
	decryptedBytes, err := domain.Decrypt(fm.authKey, res.MessageKey, res.Payload)

	receivedEncryptedPayload := new(msg.ProtoEncryptedPayload)
	err = receivedEncryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return receivedEncryptedPayload.Envelope, nil
}
