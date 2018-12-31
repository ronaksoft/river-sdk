package scenario

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"io/ioutil"
	"math/big"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"github.com/monnand/dhkx"
)

const (
	// ServerKeysFilePath server public key
	ServerKeysFilePath = "keys.json"
)

// CreateAuthKey scenario
type CreateAuthKey struct {
	Scenario
	ServerKeys *shared.ServerKeys
}

// NewCreateAuthKey initiate CreateAuthKey scenario
func NewCreateAuthKey() *CreateAuthKey {

	s := new(CreateAuthKey)
	s.ServerKeys = &shared.ServerKeys{
		DHGroups:   make([]shared.DHGroup, 0),
		PublicKeys: make([]shared.PublicKey, 0),
	}

	// Initialize Server Keys
	jsonBytes, err := ioutil.ReadFile(ServerKeysFilePath)
	if err != nil {
		panic(err)
	}
	err = s.ServerKeys.UnmarshalJSON(jsonBytes)
	if err != nil {
		panic(err)
	}
	return s
}

// Play execute CreateAuthKey scenario
func (s *CreateAuthKey) Play(act shared.Acter) {
	s.wait.Add(1)
	act.ExecuteRequest(s.initConnect(act))
}

// Step : 1
func (s *CreateAuthKey) initConnect(act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := InitConnect()

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
		s.failed("initConnect() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if s.isErrorResponse(resp) {
			return
		}
		if resp.Constructor == msg.C_InitResponse {
			x := new(msg.InitResponse)
			x.Unmarshal(resp.Message)

			// chain next request here
			act.ExecuteRequest(s.initCompleteAuth(x, act))

		} else {
			// TODO : Reporter failed
			s.failed("initConnect() successCB response type is not InitResponse")
		}
	}

	return reqEnv, successCB, timeoutCB
}

// Step : 2
func (s *CreateAuthKey) initCompleteAuth(resp *msg.InitResponse, act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {

	clientNonce := resp.ClientNonce
	serverNonce := resp.ServerNonce
	serverPubFP := resp.RSAPubKeyFingerPrint
	serverDHFP := resp.DHGroupFingerPrint
	serverPQ := resp.PQ

	// Generate DH Pub Key
	dhGroup, err := s.ServerKeys.GetDhGroup(int64(serverDHFP))
	if err != nil {
		// TODO : Reporter failed
	}
	dhPrime := big.NewInt(0)
	dhPrime.SetString(dhGroup.Prime, 16)
	dh := dhkx.CreateGroup(dhPrime, big.NewInt(int64(dhGroup.Gen)))
	dhPubKey, _ := dh.GeneratePrivateKey(rand.Reader)
	pp, qq := domain.SplitPQ(big.NewInt(int64(serverPQ)))
	var p, q uint64
	if pp.Cmp(qq) < 0 {
		p, q = pp.Uint64(), qq.Uint64()
	} else {
		p, q = qq.Uint64(), pp.Uint64()
	}

	q2Internal := new(msg.InitCompleteAuthInternal)
	q2Internal.SecretNonce = []byte(domain.RandomID(16))

	serverPubKey, err := s.ServerKeys.GetPublicKey(int64(serverPubFP))
	if err != nil {
		// TODO : Reporter failed
		s.failed("ServerKeys.GetPublicKey(), Err : " + err.Error())
	}
	n := big.NewInt(0)
	n.SetString(serverPubKey.N, 10)
	rsaPublicKey := rsa.PublicKey{
		N: n,
		E: int(serverPubKey.E),
	}
	decrypted, _ := q2Internal.Marshal()
	encPayload, err := rsa.EncryptPKCS1v15(rand.Reader, &rsaPublicKey, decrypted)
	if err != nil {
		// TODO : Reporter failed
		s.failed("rsa.EncryptPKCS1v15(), Err : " + err.Error())
	}

	// send chained request
	reqEnv := InitCompleteAuth(clientNonce, serverNonce, p, q, dhPubKey.Bytes(), encPayload)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
		s.failed("initCompleteAuth() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if s.isErrorResponse(resp) {
			return
		}
		// TODO : chain next request here
		if resp.Constructor == msg.C_InitAuthCompleted {
			x := new(msg.InitAuthCompleted)
			x.Unmarshal(resp.Message)

			switch x.Status {
			case msg.InitAuthCompleted_OK:
				serverDhKey, err := dh.ComputeKey(dhkx.NewPublicKey(x.ServerDHPubKey), dhPubKey)
				if err != nil {
					return
				}

				authKey := serverDhKey.Bytes()
				authKeyHash, _ := domain.Sha256(authKey)
				authID := int64(binary.LittleEndian.Uint64(authKeyHash[24:32]))

				// TODO : Complete Scenario
				// Save authKey && authID
				act.SetAuthInfo(authID, authKey)

				var secret []byte
				secret = append(secret, q2Internal.SecretNonce...)
				secret = append(secret, byte(msg.InitAuthCompleted_OK))
				secret = append(secret, authKeyHash[:8]...)
				secretHash, _ := domain.Sha256(secret)

				if x.SecretHash != binary.LittleEndian.Uint64(secretHash[24:32]) {
					err = domain.ErrSecretNonceMismatch
					// TODO : Reporter failed
					s.failed("initCompleteAuth(), err : " + err.Error())
					return
				}
				s.completed("initCompleteAuth() Success")
			case msg.InitAuthCompleted_RETRY:
				// TODO : Reporter failed && Retry with new DHKey
				s.failed("initCompleteAuth(), err : Retry with new DHKey")

			case msg.InitAuthCompleted_FAIL:
				err = domain.ErrAuthFailed
				// TODO : Reporter failed
				s.failed("initCompleteAuth(), err : " + err.Error())
			}

		} else {
			// TODO : Reporter failed
			s.failed("initCompleteAuth() successCB response type is not InitAuthCompleted")
		}
	}

	return reqEnv, successCB, timeoutCB
}
