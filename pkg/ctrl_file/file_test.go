package fileCtrl_test

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/valyala/tcplisten"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

/*
   Creation Time: 2019 - Sep - 08
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var (
	_File    *fileCtrl.Controller
	_Network *networkCtrl.Controller
)

func init() {
	repo.InitRepo("./db", true)
	_Network = networkCtrl.New(networkCtrl.Config{
		WebsocketEndpoint: "",
		HttpEndpoint:      "http://127.0.0.1:8080",
	})
	_File = fileCtrl.New(fileCtrl.Config{
		Network:              _Network,
		MaxInflightDownloads: 2,
		MaxInflightUploads:   10,
	})

}

type TestServer struct {
}

func (t TestServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// time.Sleep(3 * time.Second)
	if ronak.RandomInt(10) < 5 {
		res.WriteHeader(http.StatusForbidden)
		return
	}
	body, _ := ioutil.ReadAll(req.Body)
	protoMessage := new(msg.ProtoMessage)
	_ = protoMessage.Unmarshal(body)
	eIn := new(msg.MessageEnvelope)
	_ = eIn.Unmarshal(protoMessage.Payload)
	switch eIn.Constructor {
	case msg.C_FileGet:
		req := new(msg.FileGet)
		_ = req.Unmarshal(eIn.Message)
		logs.Info("Request Received:",
			zap.Int64("AuthID", protoMessage.AuthID),
			zap.String("Constructor", msg.ConstructorNames[eIn.Constructor]),
			zap.Int32("Offset", req.Offset),
			zap.Int32("Limit", req.Limit),
		)
		file := new(msg.File)
		file.Bytes = make([]byte, req.Limit)
		for i := int32(0); i < req.Limit-2; i++ {
			file.Bytes[i] = 'X'
		}
		file.Bytes[req.Limit-2] = '\r'
		file.Bytes[req.Limit-1] = '\n'
		eIn.Constructor = msg.C_File
		eIn.Message, _ = file.Marshal()
		protoMessage.Payload, _ = eIn.Marshal()
		b, _ := protoMessage.Marshal()
		_, _ = res.Write(b)

	case msg.C_FileSavePart:

	}
}

func TestDownload(t *testing.T) {
	var err error
	tcpConfig := new(tcplisten.Config)
	s := httptest.NewUnstartedServer(TestServer{})
	s.Listener, err = tcpConfig.NewListener("tcp4", ":8080")
	if err != nil {
		logs.Fatal(err.Error())
	}
	s.Start()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			_File.Download(fileCtrl.DownloadRequest{
				MaxRetries:      10,
				MessageID:       1000,
				ClusterID:       11,
				FileID:          111,
				AccessHash:      1111,
				Version:         1,
				FileSize:        2560,
				ChunkSize:       256,
				MaxInFlights:    3,
				FilePath:        fmt.Sprintf("./FILE_%d", i),
				DownloadedParts: nil,
				TotalParts:      0,
				Status:          0,
			})
			wg.Done()
		}(i)
	}
	wg.Wait()
}
