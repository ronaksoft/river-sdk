package fileCtrl_test

import (
	"crypto/md5"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
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

	waitMapLock      = sync.Mutex{}
	waitMap          = make(map[string]struct{})
	speedBytesPerSec = 1024 * 1024
	errRatePercent   int
)

func init() {
	repo.MustInit("./_hdd/_db", true)
	repo.SetRootFolders(
		"_hdd/_data/audio", "_hdd/_data/file", "_hdd/_data/photo",
		"_hdd/_data/video", "_hdd/_data/cache",
	)
	_Network = networkCtrl.New(networkCtrl.Config{
		WebsocketEndpoint: "",
		HttpEndpoint:      "http://127.0.0.1:8080",
		HttpTimeout:       10 * time.Second,
	})
	_File = fileCtrl.New(fileCtrl.Config{
		Network:              _Network,
		MaxInflightDownloads: 2,
		MaxInflightUploads:   3,
		DbPath:               "./_hdd",
		ProgressChangedCB: func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64) {
			logs.Info("Progress Changed", zap.String("ReqID", reqID), zap.Int64("Percent", percent))
		},
		CancelCB: func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64) {
			logs.Error("File Canceled", zap.String("ReqID", reqID), zap.Bool("HasError", hasError))
			waitMapLock.Lock()
			delete(waitMap, fmt.Sprintf("%d.%d.%d", clusterID, fileID, accessHash))
			waitMapLock.Unlock()
		},
		CompletedCB: func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64) {
			logs.Info("OnComplete", zap.Int64("FileID", fileID))
			waitMapLock.Lock()
			delete(waitMap, fmt.Sprintf("%d.%d.%d", clusterID, fileID, accessHash))
			waitMapLock.Unlock()
		},
		PostUploadProcessCB: func(req *msg.ClientFileRequest) bool {
			logs.Info("PostProcess",
				zap.Any("TotalParts", req.TotalParts),
				zap.Int32("ChunkSize", req.ChunkSize),
				zap.Any("FilePath", req.FilePath),
				zap.Any("FileID", req.FileID),
				zap.Int64("ThumbID", req.ThumbID),
			)
			waitMapLock.Lock()
			delete(waitMap, fmt.Sprintf("%d.%d.%d", req.ClusterID, req.FileID, req.AccessHash))
			waitMapLock.Unlock()
			return true

		},
	})
	_File.Start()
	logs.SetLogLevel(0)

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
	time.Sleep(time.Second * 2)
}

type server struct {
	mtx           *sync.Mutex
	uploadTracker map[int64]map[int32]struct{}
	sha           map[int64]hash.Hash
}

func (t server) ServeHTTP(httpRes http.ResponseWriter, httpReq *http.Request) {
	body, _ := ioutil.ReadAll(httpReq.Body)

	time.Sleep(time.Duration(len(body)/(speedBytesPerSec*(tools.RandomInt(10)+1))) * time.Second)

	if domain.RandomInt(100) > (100 - errRatePercent) {
		httpRes.WriteHeader(http.StatusForbidden)
		return
	}

	protoIn := &msg.ProtoMessage{}
	protoOut := &msg.ProtoMessage{}
	in := &rony.MessageEnvelope{}
	out := &rony.MessageEnvelope{}

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
				out.Constructor = rony.C_Error
				out.Message, _ = (&rony.Error{
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
				c.So(filePath, ShouldEqual, repo.Files.GetFilePath(clientFile))
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
	Convey("Upload", t, func(c C) {
		fileID := domain.RandomInt63()
		msgID := domain.RandomInt63()
		peerID := domain.RandomInt63()
		Convey("Good Network", func(c C) {
			Convey("Upload Big File (Good Network)", func(c C) {
				c.Println()
				waitMapLock.Lock()
				waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
				waitMapLock.Unlock()
				speedBytesPerSec = 1024 * 512
				_File.UploadMessageDocument(msgID, "./testdata/big", "", fileID, 0, nil, peerID, true)
			})
			Convey("Upload Medium File (Good Network)", func(c C) {
				c.Println()
				waitMapLock.Lock()
				waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
				waitMapLock.Unlock()
				speedBytesPerSec = 1024 * 512
				_File.UploadMessageDocument(msgID, "./testdata/medium", "", fileID, 0, nil, peerID, true)
			})
			for {
				if len(waitMap) == 0 {
					break
				}
				time.Sleep(time.Second)
			}
		})
		Convey("Bad Network", func(c C) {
			Convey("Upload Big File (Bad Network)", func(c C) {
				c.Println()
				speedBytesPerSec = 1024 * 8
				errRatePercent = 0
				waitMapLock.Lock()
				waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
				waitMapLock.Unlock()
				_File.UploadMessageDocument(msgID, "./testdata/big", "", fileID, 0, nil, peerID, true)
			})
			Convey("Upload Medium File (Bad Network)", func(c C) {
				c.Println()
				speedBytesPerSec = 8192
				errRatePercent = 0
				waitMapLock.Lock()
				waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
				waitMapLock.Unlock()
				_File.UploadMessageDocument(msgID, "./testdata/medium", "", fileID, 0, nil, peerID, true)
			})
			for {
				if len(waitMap) == 0 {
					break
				}
				time.Sleep(time.Second)
			}
		})
	})
}

func TestManyUpload(t *testing.T) {
	Convey("Upload Many Files (Good Network)", t, func(c C) {
		c.Println()
		speedBytesPerSec = 1024 * 1024
		for i := 0; i < 20; i++ {
			fileID := int64(i + 1)
			thumbID := domain.RandomInt63()
			msgID := int64(i + 1)
			peerID := int64(i + 1)
			waitMapLock.Lock()
			waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
			waitMapLock.Unlock()
			_File.UploadMessageDocument(msgID, "./testdata/big", "./testdata/small", fileID, thumbID, nil, peerID, false)
		}
		for {
			logs.Info("WaitMap", zap.Int("Size", len(waitMap)))
			if len(waitMap) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
	})
	Convey("Upload Many Files (Bad Network)", t, func(c C) {
		c.Println()
		speedBytesPerSec = 1024 * 128
		for i := 0; i < 10; i++ {
			fileID := int64(i + 1)
			msgID := int64(i + 1)
			peerID := int64(i + 1)
			waitMapLock.Lock()
			waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
			waitMapLock.Unlock()
			_File.UploadMessageDocument(msgID, "./testdata/big", "", fileID, 0, nil, peerID, true)
		}
		for {
			logs.Info("WaitMap", zap.Int("Size", len(waitMap)))
			if len(waitMap) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
	})

}

func TestUploadWithThumbnail(t *testing.T) {
	fileID := domain.RandomInt63()
	thumbID := domain.RandomInt63()
	msgID := domain.RandomInt63()
	peerID := domain.RandomInt63()
	Convey("Upload File With Thumbnail", t, func(c C) {
		c.Println()
		speedBytesPerSec = 1024 * 256
		waitMapLock.Lock()
		waitMap[fmt.Sprintf("%d.%d.%d", 0, fileID, 0)] = struct{}{}
		waitMapLock.Unlock()
		_File.UploadMessageDocument(msgID, "./testdata/small", "./testdata/small", fileID, thumbID, nil, peerID, true)
	})
	for {
		if len(waitMap) == 0 {
			break
		}
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second * 10)

}
