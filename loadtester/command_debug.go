package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"github.com/kr/pretty"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

func init() {
	cmdDebug.AddCmd(Decrypt)
	cmdDebug.AddCmd(AuthInfo)
}

var cmdDebug = &ishell.Cmd{
	Name: "Debug",
}

var Decrypt = &ishell.Cmd{
	Name: "Decrypt",
	Func: func(c *ishell.Context) {
		c.Print("Actor Phone : ")
		phone := c.ReadLine()

		c.Print("Enter Hex string : ")
		hexStr := c.ReadLine()
		rawBytes, err := hex.DecodeString(hexStr)
		if err != nil {
			logs.Error("hex.DecodeString()", zap.Error(err))
			return
		}

		act, err := actor.NewActor(phone)
		if err != nil {
			logs.Error("actor.NewActor()", zap.Error(err))
			return
		}

		authID, authKey := act.GetAuthInfo()
		if authID == 0 || len(authKey) == 0 {
			logs.Error("actor.GetAuthInfo()", zap.String("Error", "authKey not created for this actor"))
			return
		}

		envelop, err := decryptProtoMessage(rawBytes, authKey)
		if err != nil {
			logs.Error("decryptProtoMessage()", zap.Error(err))
			return
		}

		logs.Info("ProtoMessage Decrypt", zap.String("Constructor", msg.ConstructorNames[envelop.Constructor]))
		switch envelop.Constructor {
		case msg.C_AuthRecall:
			fnPrintAuthRecal(envelop)
		case msg.C_AuthRecalled:
			fnPrintAuthRecalled(envelop)
		case msg.C_ContactsImport:
			fnPrintContactsImport(envelop)
		case msg.C_ContactsImported:
			fnPrintContactsImported(envelop)
		case msg.C_Error:
			fnPrintError(envelop)
		}

	},
}

var AuthInfo = &ishell.Cmd{
	Name: "AuthInfo",
	Func: func(c *ishell.Context) {
		c.Print("Actor Phone : ")
		phone := c.ReadLine()

		act, err := actor.NewActor(phone)
		if err != nil {
			logs.Error("actor.NewActor()", zap.Error(err))
			return
		}

		authID, authKey := act.GetAuthInfo()
		if authID == 0 || len(authKey) == 0 {
			logs.Error("actor.GetAuthInfo()", zap.String("Error", "authKey not created for this actor"))
			return
		}
		hexStr := hex.EncodeToString(authKey)

		authKeyHash, _ := domain.Sha256(authKey)
		authIDBySha256OfAuthKey := int64(binary.LittleEndian.Uint64(authKeyHash[24:32]))

		logs.Info("AuthInfo",
			zap.Int64("AuthID", authID),
			zap.Int64("Sha256(authKey)[24:32]", authIDBySha256OfAuthKey),
			zap.String("AuthKey", hexStr),
		)

	},
}

func decryptProtoMessage(rawBytes, authKey []byte) (*msg.MessageEnvelope, error) {
	protMsg := new(msg.ProtoMessage)
	err := protMsg.Unmarshal(rawBytes)
	if err != nil {
		return nil, err
	}

	if protMsg.AuthID == 0 {
		env := new(msg.MessageEnvelope)
		err := env.Unmarshal(protMsg.Payload)
		if err != nil {
			return nil, err
		}
		return env, nil
	}
	decryptedBytes, err := domain.Decrypt(authKey, protMsg.MessageKey, protMsg.Payload)
	if err != nil {
		return nil, err
	}
	encryptedPayload := new(msg.ProtoEncryptedPayload)
	err = encryptedPayload.Unmarshal(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return encryptedPayload.Envelope, nil

}

func fnPrintAuthRecal(env *msg.MessageEnvelope) {
	x := new(msg.AuthRecall)
	err := x.Unmarshal(env.Message)
	if err != nil {
		logs.Error("Error", zap.Error(err))
		return
	}
	fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))
}
func fnPrintAuthRecalled(env *msg.MessageEnvelope) {
	x := new(msg.AuthRecalled)
	err := x.Unmarshal(env.Message)
	if err != nil {
		logs.Error("Error", zap.Error(err))
		return
	}
	fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))
}

func fnPrintContactsImport(env *msg.MessageEnvelope) {
	x := new(msg.ContactsImport)
	err := x.Unmarshal(env.Message)
	if err != nil {
		logs.Error("Error", zap.Error(err))
		return
	}
	fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))
}
func fnPrintContactsImported(env *msg.MessageEnvelope) {
	x := new(msg.ContactsImported)
	err := x.Unmarshal(env.Message)
	if err != nil {
		logs.Error("Error", zap.Error(err))
		return
	}
	fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))
}

func fnPrintError(env *msg.MessageEnvelope) {
	x := new(msg.ContactsImported)
	err := x.Unmarshal(env.Message)
	if err != nil {
		logs.Error("Error", zap.Error(err))
		return
	}
	fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))
}
