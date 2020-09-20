package fileCtrl

import (
	"context"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/pkg/ctrl_file/executor"
	"git.ronaksoft.com/river/sdk/pkg/domain"
	"git.ronaksoft.com/river/sdk/pkg/repo"
	"go.uber.org/zap"
	"os"
	"sync"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type DownloadRequest struct {
	msg.ClientFileRequest
	ctrl         *Controller
	mtx          sync.Mutex
	file         *os.File
	parts        chan int32
	done         chan struct{}
	lastProgress int64
}

func (d *DownloadRequest) generateFileGet(offset, limit int32) *msg.MessageEnvelope {
	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		ClusterID:  d.ClusterID,
		FileID:     d.FileID,
		AccessHash: d.AccessHash,
		Version:    d.Version,
	}
	req.Offset = offset
	req.Limit = limit

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	return envelop
}

func (d *DownloadRequest) isDownloaded(partIndex int32) bool {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	for _, index := range d.FinishedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (d *DownloadRequest) addToDownloaded(partIndex int32) {
	d.mtx.Lock()
	d.FinishedParts = append(d.FinishedParts, partIndex)
	progress := int64(float64(len(d.FinishedParts)) / float64(d.TotalParts) * 100)
	skipOnProgress := false
	if d.lastProgress > progress {
		skipOnProgress = true
	} else {
		d.lastProgress = progress
	}
	d.mtx.Unlock()
	_ = repo.Files.SaveFileRequest(d.GetID(), &d.ClientFileRequest)

	if !d.SkipDelegateCall && !skipOnProgress {
		d.ctrl.onProgressChanged(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), progress, d.PeerID)
	}

}

func (d *DownloadRequest) GetID() string {
	return getRequestID(d.ClusterID, d.FileID, d.AccessHash)
}

func (d *DownloadRequest) Prepare() error {
	_, err := os.Stat(d.TempPath)
	if err != nil {
		if os.IsNotExist(err) {
			d.file, err = os.Create(d.TempPath)
			if err != nil {
				d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
				return err
			}
		} else {
			d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			return err
		}
	} else {
		d.file, err = os.OpenFile(d.TempPath, os.O_RDWR, 0666)
		if err != nil {
			d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			return err
		}
	}

	if d.FileSize > 0 {
		err := os.Truncate(d.TempPath, d.FileSize)
		if err != nil {
			d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			return err
		}
		dividend := int32(d.FileSize / int64(d.ChunkSize))
		if d.FileSize%int64(d.ChunkSize) > 0 {
			d.TotalParts = dividend + 1
		} else {
			d.TotalParts = dividend
		}
	} else {
		d.TotalParts = 1
		d.ChunkSize = 0
	}

	// Reset FinishedParts if all parts are finished. Probably something went wrong, it is better to retry
	if int32(len(d.FinishedParts)) == d.TotalParts {
		d.FinishedParts = d.FinishedParts[:0]
	}

	// Prepare Channels to active the system dynamics
	d.parts = make(chan int32, d.TotalParts)
	d.done = make(chan struct{}, 1)
	for partIndex := int32(0); partIndex < d.TotalParts; partIndex++ {
		if d.isDownloaded(partIndex) {
			continue
		}
		d.parts <- partIndex
	}

	logs.Debug("Download Prepared",
		zap.String("ID", d.GetID()),
		zap.Int32("TotalParts", d.TotalParts),
		zap.Int32s("Finished", d.FinishedParts),
	)
	return nil
}

func (d *DownloadRequest) NextAction() executor.Action {
	select {
	case partID := <-d.parts:
		return &DownloadAction{
			id:  partID,
			req: d,
		}
	case <-d.done:
		return nil
	}
}

func (d *DownloadRequest) ActionDone(id int32) {
	finished := int32(len(d.FinishedParts)) == d.TotalParts
	if finished {
		d.done <- struct{}{}
		_ = d.file.Close()
		err := os.Rename(d.TempPath, d.FilePath)
		if err != nil {
			_ = os.Remove(d.TempPath)
			if !d.SkipDelegateCall {
				d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			}
			return
		}
		if !d.SkipDelegateCall {
			d.ctrl.onCompleted(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), d.FilePath, d.PeerID)
		}
		_ = repo.Files.DeleteFileRequest(d.GetID())
	}

}

func (d *DownloadRequest) Serialize() []byte {
	b, err := d.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

func (d *DownloadRequest) Next() executor.Request {
	return nil
}

type DownloadAction struct {
	id  int32
	req *DownloadRequest
}

func (a *DownloadAction) ID() int32 {
	return a.id
}

func (a *DownloadAction) Do(ctx context.Context) {
	offset := a.id * a.req.ChunkSize
	res, err := a.req.ctrl.network.SendHttp(ctx, a.req.generateFileGet(offset, a.req.ChunkSize))
	if err != nil {
		a.req.parts <- a.id
		return
	}
	switch res.Constructor {
	case msg.C_File:
		file := new(msg.File)
		err = file.Unmarshal(res.Message)
		if err != nil {
			logs.Warn("Downloader couldn't unmarshal server response FileGet (File)",
				zap.Error(err),
				zap.Int32("Offset", offset),
				zap.Int("Byte", len(file.Bytes)),
			)
			a.req.parts <- a.id
			return
		}
		_, err := a.req.file.WriteAt(file.Bytes, int64(offset))
		if err != nil {
			logs.Error("Downloader couldn't write to file, will retry...",
				zap.Error(err),
				zap.Int32("Offset", offset),
				zap.Int("Byte", len(file.Bytes)),
			)
			a.req.parts <- a.id
			return
		}
		a.req.addToDownloaded(a.id)
	default:
		a.req.parts <- a.id
		return
	}
}
