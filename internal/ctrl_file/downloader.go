package fileCtrl

import (
	"context"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/ctrl_file/executor"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
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
	ctrl     *Controller
	mtx      sync.Mutex
	file     *os.File
	parts    chan int32
	done     chan struct{}
	progress int64
	finished bool
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
	if d.progress > progress {
		skipOnProgress = true
	} else {
		d.progress = progress
	}
	d.mtx.Unlock()
	saved, _ := repo.Files.SaveFileRequest(d.GetID(), &d.ClientFileRequest, true)
	if saved && !d.SkipDelegateCall && !skipOnProgress {
		d.ctrl.onProgressChanged(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), progress, d.PeerID)
	}

}

func (d *DownloadRequest) cancel(err error) {
	if !d.SkipDelegateCall {
		d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), err != nil, d.PeerID)
	}
	_ = repo.Files.DeleteFileRequest(d.GetID())
}

func (d *DownloadRequest) complete() {
	if !d.SkipDelegateCall {
		d.ctrl.onCompleted(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), d.FilePath, d.PeerID)
	}
	_ = repo.Files.DeleteFileRequest(d.GetID())
}

func (d *DownloadRequest) GetID() string {
	return getRequestID(d.ClusterID, d.FileID, d.AccessHash)
}

func (d *DownloadRequest) Prepare() error {
	logger.Info("prepare DownloadRequest", zap.String("ReqID", d.GetID()))
	// Check temp file stat and if it does not exists, we create it
	_, err := os.Stat(d.TempPath)
	if err != nil {
		if os.IsNotExist(err) {
			d.file, err = os.Create(d.TempPath)
			if err != nil {
				d.cancel(err)
				return err
			}
		} else {
			d.cancel(err)
			return err
		}
	} else {
		d.file, err = os.OpenFile(d.TempPath, os.O_RDWR, 0666)
		if err != nil {
			d.cancel(err)
			return err
		}
	}

	// If the size is known, we truncate the temp file
	if d.FileSize > 0 {
		err := os.Truncate(d.TempPath, d.FileSize)
		if err != nil {
			d.cancel(err)
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

	logger.Debug("Download Prepared",
		zap.String("ID", d.GetID()),
		zap.Int32("TotalParts", d.TotalParts),
		zap.Int32s("Finished", d.FinishedParts),
	)
	return nil
}

func (d *DownloadRequest) NextAction() executor.Action {
	// If request is canceled then return nil
	if _, err := repo.Files.GetFileRequest(d.GetID()); err != nil {
		logger.Warn("did not find DownloadRequest, we cancel it", zap.Error(err))
		return nil
	}

	// Wait for next part, or return nil if we finished
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
	logger.Info("finished download part",
		zap.String("ReqID", d.GetID()),
		zap.Int32("PartID", id),
		zap.Int("FinishedParts", len(d.FinishedParts)),
		zap.Int32("TotalParts", d.TotalParts),
	)

	if int32(len(d.FinishedParts)) == d.TotalParts {
		if d.finished {
			return
		}
		d.finished = true
		d.done <- struct{}{}
		_ = d.file.Close()
		err := os.Rename(d.TempPath, d.FilePath)
		if err != nil {
			_ = os.Remove(d.TempPath)
			d.cancel(err)
			return
		}
		d.complete()
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
	// We don't support chained downloads, no need so far
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
	ctx, cf := context.WithTimeout(ctx, domain.HttpRequestTimeout)
	defer cf()
	req := &msg.FileGet{
		Location: &msg.InputFileLocation{
			ClusterID:  a.req.ClusterID,
			FileID:     a.req.FileID,
			AccessHash: a.req.AccessHash,
			Version:    a.req.Version,
		},
		Offset: offset,
		Limit:  a.req.ChunkSize,
	}

	reqCB := request.NewCallback(
		0, 0, domain.NextRequestID(), msg.C_FileGet, req,
		func() {
			a.req.parts <- a.id
		},
		func(res *rony.MessageEnvelope) {
			switch res.Constructor {
			case msg.C_File:
				file := &msg.File{}
				err := file.Unmarshal(res.Message)
				if err != nil {
					logger.Warn("couldn't unmarshal server response FileGet (File), will retry ...",
						zap.Error(err),
						zap.Int32("Offset", offset),
						zap.Int("Byte", len(file.Bytes)),
					)
					a.req.parts <- a.id
					return
				}
				_, err = a.req.file.WriteAt(file.Bytes, int64(offset))
				if err != nil {
					logger.Error("couldn't write to file, will retry...",
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
		}, nil, false, 0, domain.HttpRequestTimeout)
	a.req.ctrl.network.HttpCommand(ctx, reqCB)
}
