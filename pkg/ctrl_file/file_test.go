package fileCtrl_test

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"github.com/valyala/tcplisten"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
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
	_File *fileCtrl.Controller
)

func init() {
	repo.InitRepo("./_db", true)
	fileCtrl.SetRootFolders("_data/audio", "_data/file", "_data/photo", "_data/video", "_data/cache")
	_File = fileCtrl.New(fileCtrl.Config{
		Network: networkCtrl.New(networkCtrl.Config{
			WebsocketEndpoint: "",
			HttpEndpoint:      "http://127.0.0.1:8080",
		}),
		MaxInflightDownloads: 2,
		MaxInflightUploads:   10,
		OnProgressChanged: func(reqID string, clusterID int32, fileID, accessHash int64, percent int64) {
			// logs.Info("Progress Changed", zap.String("ReqID", reqID), zap.Int64("Percent", percent))
		},
		OnCancel: func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool) {
			logs.Error("File Canceled", zap.String("ReqID", reqID), zap.Bool("HasError", hasError))
		},
		OnCompleted: func(reqID string, clusterID int32, fileID, accessHash int64, filePath string) {
			logs.Info("File Completed", zap.String("ReqID", reqID))
		},
		PostUploadProcess: func(req fileCtrl.UploadRequest) {
			logs.Info("PostProcess:", zap.Any("TotalParts", req.TotalParts))
		},
	})
	_File.Start()

	tcpConfig := new(tcplisten.Config)
	s := httptest.NewUnstartedServer(server{
		uploadTracker: make(map[int64]map[int32]struct{}),
	})

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			err := repo.Files.Save(&msg.ClientFile{
				ClusterID:  1,
				FileID:     int64(i),
				AccessHash: 10,
				Type:       msg.ClientFileType_Message,
				MimeType:   "video/mp4",
				UserID:     0,
				GroupID:    0,
				FileSize:   1024000,
				MessageID:  int64(i),
				PeerID:     0,
				PeerType:   0,
				Version:    0,
			})
			if err != nil {
				panic(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	var err error
	s.Listener, err = tcpConfig.NewListener("tcp4", ":8080")
	if err != nil {
		logs.Fatal(err.Error())
	}
	s.Start()

}

type server struct {
	sync.Mutex
	uploadTracker map[int64]map[int32]struct{}
}

func (t server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// time.Sleep(3 * time.Second)
	// if ronak.RandomInt(30) > 5 {
	// 	res.WriteHeader(http.StatusForbidden)
	// 	return
	// }
	body, _ := ioutil.ReadAll(req.Body)
	protoMessage := new(msg.ProtoMessage)
	_ = protoMessage.Unmarshal(body)
	eIn := new(msg.MessageEnvelope)
	_ = eIn.Unmarshal(protoMessage.Payload)
	switch eIn.Constructor {
	case msg.C_FileGet:
		req := new(msg.FileGet)
		_ = req.Unmarshal(eIn.Message)
		logs.Info("FileGet",
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
		req := new(msg.FileSavePart)
		_ = req.Unmarshal(eIn.Message)
		logs.Info("SavePart:",
			zap.Int64("FileID", req.FileID),
			zap.Int32("PartID", req.PartID),
			zap.Int32("TotalParts", req.TotalParts),
		)
		t.Lock()
		if _, ok := t.uploadTracker[req.FileID]; !ok {
			t.uploadTracker[req.FileID] = make(map[int32]struct{})
		}
		t.uploadTracker[req.FileID][req.PartID] = struct{}{}
		t.Unlock()
		if req.PartID == req.TotalParts {
			sum := int32(0)
			t.Lock()
			for partID := range t.uploadTracker[req.FileID] {
				sum += partID
			}
			t.Unlock()
			if sum == (req.TotalParts*(req.TotalParts+1))/2 {
				logs.Info("CORRECT UPLOAD")
			}
		}
		eIn.Constructor = msg.C_Bool
		eIn.Message, _ = (&msg.Bool{}).Marshal()
		protoMessage.Payload, _ = eIn.Marshal()
		b, _ := protoMessage.Marshal()
		_, _ = res.Write(b)
	}
}

func TestDownloadFileSync(t *testing.T) {
	wg := sync.WaitGroup{}

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			clientFile, err := repo.Files.Get(1, int64(i), 10)
			if err != nil {
				t.Fatal(err)
			}
			filePath, err := _File.DownloadSync(clientFile.ClusterID, clientFile.FileID, clientFile.AccessHash, false)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(i, filePath)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestDownloadFileASync(t *testing.T) {
	for i := 0; i < 1; i++ {
		clientFile, err := repo.Files.Get(1, int64(i), 10)
		if err != nil {
			t.Fatal(err)
		}
		filePath, err := _File.DownloadAsync(clientFile.ClusterID, clientFile.FileID, clientFile.AccessHash, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(i, filePath)
	}
	time.Sleep(10 * time.Second)

}

func TestUpload(t *testing.T) {
	// _File.UploadUserPhoto("./testdata/big")
	_File.UploadMessageDocument(1000, "./testdata/big", "", 323232, 10)
}
