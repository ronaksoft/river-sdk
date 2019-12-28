package fileCtrl_test

import (
	"crypto/md5"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/valyala/tcplisten"
	"go.uber.org/zap"
	"hash"
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
	_Network *networkCtrl.Controller
	_File    *fileCtrl.Controller

	uploadStart      = false
	waitGroupUpload  sync.WaitGroup
	speedBytesPerSec int = 1024 * 1024
	errRatePercent   int
)

func init() {
	repo.InitRepo("./_db", true)
	fileCtrl.SetRootFolders("_data/audio", "_data/file", "_data/photo", "_data/video", "_data/cache")
	_Network = networkCtrl.New(networkCtrl.Config{
		WebsocketEndpoint: "",
		HttpEndpoint:      "http://127.0.0.1:8080",
	})
	_File = fileCtrl.New(fileCtrl.Config{
		Network:              _Network,
		MaxInflightDownloads: 2,
		MaxInflightUploads:   10,
		OnProgressChanged: func(reqID string, clusterID int32, fileID, accessHash int64, percent int64) {
			// logs.Info("Progress Changed", zap.String("ReqID", reqID), zap.Int64("Percent", percent))
		},
		OnCancel: func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool) {
			logs.Error("File Canceled", zap.String("ReqID", reqID), zap.Bool("HasError", hasError))
			if clusterID == 0 && uploadStart {
				// It is an Upload
				waitGroupUpload.Done()
			}
		},
		OnCompleted: func(reqID string, clusterID int32, fileID, accessHash int64, filePath string) {},
		PostUploadProcess: func(req fileCtrl.UploadRequest) {
			logs.Info("PostProcess",
				zap.Any("TotalParts", req.TotalParts),
				zap.Any("FilePath", req.FilePath),
				zap.Any("FileID", req.FileID),
			)
			if uploadStart {
				waitGroupUpload.Done()
			}

		},
	})
	_File.Start()

	tcpConfig := new(tcplisten.Config)
	s := httptest.NewUnstartedServer(server{
		mtx:           &sync.Mutex{},
		uploadTracker: make(map[int64]map[int32]struct{}),
		sha:           make(map[int64]hash.Hash),
	})

	wg := sync.WaitGroup{}

	for i := 1; i < 10; i++ {
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
	mtx           *sync.Mutex
	uploadTracker map[int64]map[int32]struct{}
	sha           map[int64]hash.Hash
}

func (t server) ServeHTTP(httpRes http.ResponseWriter, httpReq *http.Request) {
	body, _ := ioutil.ReadAll(httpReq.Body)
	time.Sleep(time.Duration(len(body) / speedBytesPerSec))

	if ronak.RandomInt(100) > (100 - errRatePercent) {
		httpRes.WriteHeader(http.StatusForbidden)
		return
	}

	protoIn := &msg.ProtoMessage{}
	protoOut := &msg.ProtoMessage{}
	in := &msg.MessageEnvelope{}
	out := &msg.MessageEnvelope{}

	_ = protoIn.Unmarshal(body)
	_ = in.Unmarshal(protoIn.Payload)
	switch in.Constructor {
	case msg.C_FileGet:
		req := &msg.FileGet{}
		_ = req.Unmarshal(in.Message)
		// logs.Info("FileGet", zap.Int32("Offset", httpReq.Offset), zap.Int32("Limit", httpReq.Limit))
		file := &msg.File{}
		file.Bytes = make([]byte, req.Limit)
		for i := int32(0); i < req.Limit-2; i++ {
			file.Bytes[i] = 'X'
		}
		file.Bytes[req.Limit-2] = '\r'
		file.Bytes[req.Limit-1] = '\n'

		out.Constructor = msg.C_File
		out.Message, _ = file.Marshal()
		protoOut.Payload, _ = out.Marshal()
		b, _ := protoOut.Marshal()
		_, _ = httpRes.Write(b)
	case msg.C_FileSavePart:
		req := &msg.FileSavePart{}
		_ = req.Unmarshal(in.Message)
		// logs.Info("SavePart:", zap.Int64("FileID", req.FileID), zap.Int32("PartID", req.PartID), zap.Int32("TotalParts", req.TotalParts))
		t.mtx.Lock()
		if _, ok := t.uploadTracker[req.FileID]; !ok {
			t.uploadTracker[req.FileID] = make(map[int32]struct{})
			t.sha[req.FileID] = md5.New()
		}
		t.uploadTracker[req.FileID][req.PartID] = struct{}{}
		t.sha[req.FileID].Write(req.Bytes)
		t.mtx.Unlock()
		if req.PartID == req.TotalParts {
			sum := int32(0)
			t.mtx.Lock()
			for partID := range t.uploadTracker[req.FileID] {
				sum += partID
			}
			t.mtx.Unlock()
			if sum != (req.TotalParts*(req.TotalParts+1))/2 {
				out.Constructor = msg.C_Error
				out.Message, _ = (&msg.Error{
					Code:  msg.ErrCodeIncomplete,
					Items: msg.ErrItemFileParts,
				}).Marshal()
				protoOut.Payload, _ = out.Marshal()
				b, _ := protoOut.Marshal()
				_, _ = httpRes.Write(b)
			}
		}
		out.Constructor = msg.C_Bool
		out.Message, _ = (&msg.Bool{}).Marshal()
		protoOut.Payload, _ = out.Marshal()
		b, _ := protoOut.Marshal()
		_, _ = httpRes.Write(b)
	}
}

func TestDownloadFileSync(t *testing.T) {
	Convey("DownloadFile (Synced)", t, func(c C) {
		wg := sync.WaitGroup{}

		for i := 1; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				clientFile, err := repo.Files.Get(1, int64(i), 10)
				c.So(err, ShouldBeNil)
				filePath, err := _File.DownloadSync(clientFile.ClusterID, clientFile.FileID, clientFile.AccessHash, false)
				if err != domain.ErrAlreadyDownloading {
					c.So(err, ShouldBeNil)
				}
				c.So(filePath, ShouldEqual, fileCtrl.GetFilePath(clientFile))
				wg.Done()
			}(i)
		}
		wg.Wait()
	})

}

func TestDownloadFileASync(t *testing.T) {
	Convey("DownloadFile (Async)", t, func(c C) {
		for i := 5; i < 10; i++ {
			clientFile, err := repo.Files.Get(1, int64(i), 10)
			c.So(err, ShouldBeNil)
			_, err = _File.DownloadAsync(clientFile.ClusterID, clientFile.FileID, clientFile.AccessHash, false)
			c.So(err, ShouldBeNil)
		}
	})

}

func TestUpload(t *testing.T) {
	uploadStart = true
	Convey("Upload", t, func() {
		fileID := ronak.RandomInt64(0)
		msgID := ronak.RandomInt64(0)
		Convey("Good Network", func(c C) {
			startTime := time.Now()
			Convey("Upload Big File (Good Network)", func(c C) {
				waitGroupUpload.Add(1)
				speedBytesPerSec = 1024 * 512
				_File.UploadMessageDocument(msgID, "./testdata/big", "", fileID, 0)
			})
			Convey("Upload Medium File (Good Network)", func(c C) {
				waitGroupUpload.Add(1)
				speedBytesPerSec = 1024 * 512
				_File.UploadMessageDocument(msgID, "./testdata/medium", "", fileID, 0)
			})

			waitGroupUpload.Wait()
			_, _ = Println("Good Network:", time.Now().Sub(startTime))
		})
		Convey("Bad Network", func(c C) {
			startTime := time.Now()
			Convey("Upload Big File (Bad Network)", func(c C) {
				speedBytesPerSec = 1024
				errRatePercent = 50
				waitGroupUpload.Add(1)
				_File.UploadMessageDocument(msgID, "./testdata/big", "", fileID, 0)
			})
			Convey("Upload Medium File (Bad Network)", func(c C) {
				speedBytesPerSec = 1024
				errRatePercent = 50
				waitGroupUpload.Add(1)
				_File.UploadMessageDocument(msgID, "./testdata/medium", "", fileID, 0)
			})
			waitGroupUpload.Wait()
			_, _ = Println("Bad Network:", time.Now().Sub(startTime))
		})
	})

}

// func TestContext(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	go func() {
// 		_, err := _Network.SendHttp(ctx, &msg.MessageEnvelope{
// 			Constructor: 0,
// 			RequestID:   0,
// 			Message:     nil,
// 		})
// 		if err != nil && err != context.Canceled {
// 			t.Error(err)
// 		}
// 	}()
// 	time.Sleep(time.Second * 3)
// 	cancel()
//
// 	time.Sleep(time.Second)
//
// }
