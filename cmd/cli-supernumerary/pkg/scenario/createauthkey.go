package scenario

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"io/ioutil"
	"math/big"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"github.com/monnand/dhkx"
)

const (
	ServerKeysFilePath = "keys.json"
)

var (
	_ServerKeys *shared.ServerKeys
)


// CreateAuthKey scenario
type CreateAuthKey struct {
	Scenario
}

// NewCreateAuthKey initiate CreateAuthKey scenario
func NewCreateAuthKey(isFinal bool) shared.Screenwriter {
	s := new(CreateAuthKey)
	s.isFinal = isFinal
	return s
}

// Play execute CreateAuthKey scenario
func (s *CreateAuthKey) Play(act shared.Actor) {
	if act.GetAuthID() != 0 {
		s.log(act, "Actor already have AuthID", 0, 0)
		return
	}
	s.AddJobs(1)
	act.ExecuteRequest(s.initConnect(act))
}

// Step : 1
func (s *CreateAuthKey) initConnect(act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := InitConnect()

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_InitConnect, elapsed)
		s.failed(act, elapsed, requestID, "initConnect() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_InitConnect, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "initConnect()") {
			return
		}
		if resp.Constructor == msg.C_InitResponse {
			x := new(msg.InitResponse)
			x.Unmarshal(resp.Message)

			// chain next request here
			act.ExecuteRequest(s.initCompleteAuth(x, act))
			s.log(act, "initConnect() Success", elapsed, resp.RequestID)

		} else {
			s.failed(act, elapsed, resp.RequestID, "initConnect() successCB response type is not InitResponse , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}

// Step : 2
func (s *CreateAuthKey) initCompleteAuth(resp *msg.InitResponse, act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	clientNonce := resp.ClientNonce
	serverNonce := resp.ServerNonce
	serverPubFP := resp.RSAPubKeyFingerPrint
	serverDHFP := resp.DHGroupFingerPrint
	serverPQ := resp.PQ

	// Generate DH Pub Key
	dhGroup, _ := _ServerKeys.GetDhGroup(int64(serverDHFP))
	dhPrime := big.NewInt(0)
	dhPrime.SetString(dhGroup.Prime, 16)

	dh := dhkx.CreateGroup(dhPrime, big.NewInt(int64(dhGroup.Gen)))
	dhPubKey, _ := dh.GeneratePrivateKey(rand.Reader)
	// dhPubKey := GetDhPrivateKey()
	// defer PutDhPrivateKey(dhPubKey)

	pp, qq := domain.SplitPQ(big.NewInt(int64(serverPQ)))
	var p, q uint64
	if pp.Cmp(qq) < 0 {
		p, q = pp.Uint64(), qq.Uint64()
	} else {
		p, q = qq.Uint64(), pp.Uint64()
	}

	q2Internal := new(msg.InitCompleteAuthInternal)
	q2Internal.SecretNonce = []byte(domain.RandomID(16))

	serverPubKey, err := _ServerKeys.GetPublicKey(int64(serverPubFP))
	if err != nil {
		s.failed(act, -1, 0, "ServerKeys.GetPublicKey(), Err : "+err.Error())
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
		s.failed(act, -1, 0, "rsa.EncryptPKCS1v15(), Err : "+err.Error())
	}

	// send chained request
	reqEnv := InitCompleteAuth(clientNonce, serverNonce, p, q, dhPubKey.Bytes(), encPayload)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_InitCompleteAuth, elapsed)
		s.failed(act, elapsed, requestID, "initCompleteAuth() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_InitCompleteAuth, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "InitCompleteAuth()") {
			return
		}

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

				var secret []byte
				secret = append(secret, q2Internal.SecretNonce...)
				secret = append(secret, byte(msg.InitAuthCompleted_OK))
				secret = append(secret, authKeyHash[:8]...)
				secretHash, _ := domain.Sha256(secret)

				clientSecretHash := binary.LittleEndian.Uint64(secretHash[24:32])
				if x.SecretHash != clientSecretHash {
					err = domain.ErrSecretNonceMismatch
					s.failed(act, elapsed, resp.RequestID, "initCompleteAuth(), err : "+err.Error())
					_Log.Error("initCompleteAuth(), err : secret hash does not match",
						zap.Uint64("Server SecretHash", x.SecretHash),
						zap.Uint64("Client SecretHash", clientSecretHash),
					)
					return
				}

				// Save authKey && authID
				act.SetAuthInfo(authID, authKey)
				if s.isFinal {
					err := act.Save()
					if err != nil {
						s.log(act, "initCompleteAuth() Actor.Save(), Err : "+err.Error(), elapsed, resp.RequestID)
					}
				}
				s.completed(act, elapsed, resp.RequestID, "initCompleteAuth() Success")
			case msg.InitAuthCompleted_RETRY:
				// retry to create auth key
				act.ExecuteRequest(s.initConnect(act))

			case msg.InitAuthCompleted_FAIL:
				err = domain.ErrAuthFailed
				s.failed(act, elapsed, resp.RequestID, "initCompleteAuth(), err : "+err.Error())
			}

		} else {
			s.failed(act, elapsed, resp.RequestID, "initCompleteAuth() successCB response type is not InitAuthCompleted , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}

const (
	dhKeys = 5000
)

type generator struct {
	dhKeyChan chan *dhkx.DHKey
}

var (
	_Generator = generator{
		dhKeyChan: make(chan *dhkx.DHKey, dhKeys),
	}
)

func LoadServerKeys() {
	_ServerKeys = &shared.ServerKeys{
		DHGroups:   make([]shared.DHGroup, 0),
		PublicKeys: make([]shared.PublicKey, 0),
	}

	// Initialize Server Keys
	jsonBytes, err := ioutil.ReadFile(ServerKeysFilePath)
	if err != nil {
		panic(err)
	}
	err = _ServerKeys.UnmarshalJSON(jsonBytes)
	if err != nil {
		panic(err)
	}
}

func GetDhPrivateKey() *dhkx.DHKey {
	return <-_Generator.dhKeyChan
}

func PutDhPrivateKey(key *dhkx.DHKey) {
	_Generator.dhKeyChan <- key
}

func GenDhPrivateKey() {
	dhGroup := _ServerKeys.DHGroups[0]
	dhPrime := big.NewInt(0)
	dhPrime.SetString(dhGroup.Prime, 16)

	dh := dhkx.CreateGroup(dhPrime, big.NewInt(int64(dhGroup.Gen)))

	n := runtime.NumCPU()
	waitGroup := sync.WaitGroup{}
	genFunc := func() {
		for {
			key, err := dh.GeneratePrivateKey(rand.Reader)
			if err != nil {
				_Log.Warn("Error On DhPrivateKey Generator")
				time.Sleep(time.Second)
			}
			select {
			case _Generator.dhKeyChan <- key:
			default:
				waitGroup.Done()
				return
			}
		}
	}

	waitGroup.Add(n)
	for i := 0; i < n; i++ {
		go genFunc()
	}
	waitGroup.Wait()
}
