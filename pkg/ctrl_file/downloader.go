package fileCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
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

type downloadJob struct {
	retries    int
	MessageID  int64
	ClusterID  int32
	FileID     int64
	AccessHash uint64
	Version    int32
	Offset     int32
	Limit      int32
}

type downloadStatus struct {
	TotalParts 	int32
	FilePath 	string

}

type downloader struct {
	jobChan     chan *downloadJob
	stopChan    chan struct{}
	started     bool
	waitGroup   sync.WaitGroup
	networkCtrl *networkCtrl.Controller
}

func newDownloader(network *networkCtrl.Controller, noWorkers int) *downloader {
	d := new(downloader)
	d.jobChan = make(chan *downloadJob, 100)
	d.stopChan = make(chan struct{}, noWorkers)
	d.networkCtrl = network
	return d
}

func (d *downloader) worker() {
	defer d.waitGroup.Done()
	for {
		select {
		case job := <-d.jobChan:
			res, err := d.networkCtrl.SendHttp(generateFileGet(job))
			if err != nil {
				if job.retries++; job.retries < domain.FileMaxRetry {
					d.jobChan <- job
				}
				break
			}
			switch res.Constructor {
			case msg.C_File:
				x := new(msg.File)
				err := x.Unmarshal(res.Message)
				if err != nil {
					logs.Error("downloadRequest() failed to unmarshal C_File", zap.Error(err))
					break
				}

				if len(x.Bytes) == 0 {
					logs.Error("downloadRequest() Received 0 bytes from server ",
						zap.Int64("MsgID", job.MessageID),
					)
					break

				} else {

				}

			default:
				if job.retries++; job.retries < domain.FileMaxRetry {
					d.jobChan <- job
				}
				break
			}
		case <-d.stopChan:
			return
		}
	}
}

func generateFileGet(job *downloadJob) *msg.MessageEnvelope {
	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		ClusterID:  job.ClusterID,
		FileID:     job.FileID,
		AccessHash: job.AccessHash,
		Version:    job.Version,
	}
	req.Offset = job.Offset
	req.Limit = job.Limit

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())
	logs.Debug("FilesStatus::generateFileGet()",
		zap.Int64("MsgID", job.MessageID),
		zap.Int32("Offset", req.Offset),
		zap.Int32("Limit", req.Limit),
		zap.Int64("FileID", req.Location.FileID),
		zap.Uint64("AccessHash", req.Location.AccessHash),
		zap.Int32("ClusterID", req.Location.ClusterID),
		zap.Int32("Version", req.Location.Version),
	)
	return envelop
}

func (d *downloader) Start() {
	for i := 0; i < cap(d.stopChan); i++ {
		d.waitGroup.Add(1)
		go d.worker()
	}
	d.started = true
}

func (d *downloader) Stop() {
	if d.started != true {
		panic("downloader stopped before started. BUG!")
	}
	for i := 0; i < cap(d.stopChan); i++ {
		d.stopChan <- struct{}{}
	}
	d.waitGroup.Wait()
}
