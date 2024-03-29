package fileCtrl

import (
    "context"
    "fmt"
    "io"
    "os"
    "sync"
    "sync/atomic"
    "time"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/ctrl_file/executor"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/pools"
    "github.com/ronaksoft/rony/registry"
    "go.uber.org/zap"
)

/*
   Creation Time: 2019 - Sep - 07
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type UploadRequest struct {
    cfr           *msg.ClientFileRequest
    ctrl          *Controller
    mtx           sync.Mutex
    file          *os.File
    parts         chan int32
    done          chan struct{}
    lastPartSent  bool
    progress      int64
    failedActions int32
    startTime     time.Time
}

func (u *UploadRequest) checkSha256() (err error) {
    req := &msg.FileGetBySha256{
        Sha256:   u.cfr.FileSha256,
        FileSize: int32(u.cfr.FileSize),
    }
    reqCB := request.NewCallback(
        0, 0, domain.NextRequestID(), msg.C_FileGetBySha256, req,
        func() {
            err = domain.ErrRequestTimeout
        },
        func(res *rony.MessageEnvelope) {
            switch res.Constructor {
            case msg.C_FileLocation:
                x := &msg.FileLocation{}
                _ = x.Unmarshal(res.Message)
                u.cfr.ClusterID = x.ClusterID
                u.cfr.AccessHash = x.AccessHash
                u.cfr.FileID = x.FileID
                u.cfr.TotalParts = -1 // dirty hack, which queue.Start() knows the upload request is completed
                return
            case rony.C_Error:
                x := &rony.Error{}
                _ = x.Unmarshal(res.Message)
                err = x
            }
        },
        nil, false, 0, domain.HttpRequestTimeout,
    )

    u.ctrl.network.HttpCommand(nil, reqCB)

    return
}

func (u *UploadRequest) resetUploadedList() {
    u.mtx.Lock()
    u.cfr.FinishedParts = u.cfr.FinishedParts[:0]
    u.mtx.Unlock()

    _, _ = repo.Files.SaveFileRequest(u.GetID(), u.cfr, false)
    if !u.cfr.SkipDelegateCall {
        u.ctrl.onProgressChanged(u.GetID(), 0, u.cfr.FileID, 0, 0, u.cfr.PeerID)
    }
}

func (u *UploadRequest) isUploaded(partIndex int32) bool {
    u.mtx.Lock()
    defer u.mtx.Unlock()
    for _, index := range u.cfr.FinishedParts {
        if partIndex == index {
            return true
        }
    }
    return false
}

func (u *UploadRequest) addToUploaded(partIndex int32) {
    if u.isUploaded(partIndex) {
        return
    }
    u.mtx.Lock()
    u.cfr.FinishedParts = append(u.cfr.FinishedParts, partIndex)
    progress := int64(float64(len(u.cfr.FinishedParts)) / float64(u.cfr.TotalParts) * 100)
    skipOnProgress := false
    if u.progress > progress {
        skipOnProgress = true
    } else {
        u.progress = progress
    }
    u.mtx.Unlock()

    saved, _ := repo.Files.SaveFileRequest(u.GetID(), u.cfr, true)
    if saved && !u.cfr.SkipDelegateCall && !skipOnProgress {
        u.ctrl.onProgressChanged(u.GetID(), 0, u.cfr.FileID, 0, progress, u.cfr.PeerID)
    }
}

func (u *UploadRequest) reset() {
    // Reset failed counter
    atomic.StoreInt32(&u.failedActions, 0)

    // Reset the uploaded list
    u.resetUploadedList()
    u.progress = 0

    if u.file != nil {
        _ = u.file.Close()
    }
}

func (u *UploadRequest) cancel(err error) {
    _ = repo.Files.DeleteFileRequest(u.GetID())
    if !u.cfr.SkipDelegateCall {
        u.ctrl.onCancel(u.GetID(), 0, u.cfr.FileID, 0, err != nil, u.cfr.PeerID)
    }
}

func (u *UploadRequest) complete() {
    _ = repo.Files.DeleteFileRequest(u.GetID())
    if !u.cfr.SkipDelegateCall {
        u.ctrl.onCompleted(u.GetID(), 0, u.cfr.FileID, 0, u.cfr.FilePath, u.cfr.PeerID)
    }

}

func (u *UploadRequest) GetID() string {
    return getRequestID(u.cfr.ClusterID, u.cfr.FileID, u.cfr.AccessHash)
}

func (u *UploadRequest) Prepare() error {
    logger.Info("prepares UploadRequest",
        zap.String("ReqID", u.GetID()),
        zap.Duration("D", domain.Now().Sub(u.startTime)),
    )
    st0 := domain.Now()
    u.reset()

    // Check File stats and return error if any problem exists
    fileInfo, err := os.Stat(u.cfr.FilePath)
    if err != nil {
        u.cancel(err)
        return err
    } else {
        u.cfr.FileSize = fileInfo.Size()
        if u.cfr.FileSize <= 0 {
            err = domain.ErrInvalidData
            u.cancel(err)
            return err
        } else if u.cfr.FileSize > MaxFileSizeAllowedSize {
            err = domain.ErrFileTooLarge
            u.cancel(err)
            return err
        }
    }

    // If Sha256 exists in the request then we check server if this file has been already uploaded, if true, then
    // we do not upload it again and we call postUploadProcess with the updated details
    st1 := domain.Now()
    if u.cfr.CheckSha256 && len(u.cfr.FileSha256) != 0 {
        oldReqID := u.GetID()
        err = u.checkSha256()
        if err == nil {
            logger.Info("detects the file already exists in the server",
                zap.Int64("FileID", u.cfr.FileID),
                zap.Int32("ClusterID", u.cfr.ClusterID),
            )
            _ = repo.Files.DeleteFileRequest(oldReqID)
            err = domain.ErrAlreadyUploaded
            if !u.ctrl.postUploadProcess(u.cfr) {
                err = domain.ErrInvalidData
                u.cancel(err)
            }
            return err
        }
    }
    st2 := domain.Now()

    // Open the file for read
    u.file, err = os.OpenFile(u.cfr.FilePath, os.O_RDONLY, 0)
    if err != nil {
        u.cancel(err)
        return err
    }

    // If chunk size is not set recalculate it
    if u.cfr.ChunkSize <= 0 {
        u.cfr.ChunkSize = bestChunkSize(u.cfr.FileSize)
    }

    // Calculate number of parts based on our chunk size
    dividend := int32(u.cfr.FileSize / int64(u.cfr.ChunkSize))
    if u.cfr.FileSize%int64(u.cfr.ChunkSize) > 0 {
        u.cfr.TotalParts = dividend + 1
    } else {
        u.cfr.TotalParts = dividend
    }

    // Reset FinishedParts if all parts are finished. Probably something went wrong, it is better to retry
    if int32(len(u.cfr.FinishedParts)) == u.cfr.TotalParts {
        u.cfr.FinishedParts = u.cfr.FinishedParts[:0]
    }

    // Prepare Channels to active the system dynamics
    u.parts = make(chan int32, u.cfr.TotalParts)
    u.done = make(chan struct{}, 1)
    maxPartIndex := u.cfr.TotalParts - 1
    if u.cfr.TotalParts == 1 {
        maxPartIndex = u.cfr.TotalParts
    }
    for partIndex := int32(0); partIndex < maxPartIndex; partIndex++ {
        if u.isUploaded(partIndex) {
            continue
        }
        u.parts <- partIndex
    }

    st3 := domain.Now()
    logger.Debug("prepared UploadRequest",
        zap.String("ReqID", u.GetID()),
        zap.Duration("D", domain.Now().Sub(u.startTime)),
        zap.Duration("CheckShaD", st2.Sub(st1)),
        zap.Duration("PrepareD", st3.Sub(st0)),
        zap.String("Progress", fmt.Sprintf("%d / %d", len(u.cfr.FinishedParts), u.cfr.TotalParts)),
    )
    return nil
}

func (u *UploadRequest) NextAction() executor.Action {
    // If request is canceled then return nil
    if _, err := repo.Files.GetFileRequest(u.GetID()); err != nil {
        logger.Warn("did not find UploadRequest, we cancel it", zap.Error(err))
        return nil
    }

    // Wait for next part, or return nil if we finished
    select {
    case partID := <-u.parts:
        logger.Debug("got next upload part",
            zap.String("ReqID", u.GetID()),
            zap.Int32("PartID", partID),
            zap.Duration("D", domain.Now().Sub(u.startTime)),
        )
        return &UploadAction{
            id:  partID,
            req: u,
        }
    case <-u.done:
        return nil
    }
}

func (u *UploadRequest) ActionDone(id int32) {
    logger.Info("finished upload part",
        zap.String("ID", u.GetID()),
        zap.Int32("PartID", id),
        zap.Duration("D", domain.Now().Sub(u.startTime)),
        zap.String("Progress", fmt.Sprintf("%d / %d", len(u.cfr.FinishedParts), u.cfr.TotalParts)),
    )
    // If we have failed too many times, and we can decrease the chunk size the we do it again.
    if atomic.LoadInt32(&u.failedActions) > RetryMaxAttempts {
        atomic.StoreInt32(&u.failedActions, 0)
        logger.Debug("Max Attempts",
            zap.Int32("ChunkSize", u.cfr.ChunkSize),
        )
    }

    // For single part uploads we are done
    // For n-part uploads if we have done n-1 part then we add the last part
    finishedParts := int32(len(u.cfr.FinishedParts))
    switch u.cfr.TotalParts {
    case 1:
        if finishedParts != u.cfr.TotalParts {
            return
        }
    default:
        switch {
        case finishedParts < u.cfr.TotalParts-1:
            return
        case finishedParts == u.cfr.TotalParts-1:
            if u.lastPartSent {
                return
            }
            u.lastPartSent = true
            u.parts <- u.cfr.TotalParts - 1
            return
        }
    }

    // This is last part so we make the executor free to run the next job if exist
    u.done <- struct{}{}
    _ = u.file.Close()

    // Run the post process
    if !u.ctrl.postUploadProcess(u.cfr) {
        u.cancel(domain.ErrNoPostProcess)
        return
    }

    // Clean up
    u.complete()
}

func (u *UploadRequest) Serialize() []byte {
    b, err := u.cfr.Marshal()
    if err != nil {
        panic(err)
    }
    return b
}

func (u *UploadRequest) Next() executor.Request {
    // If the request has a chained request then we swap them and reset the progress
    if u.cfr.Next == nil {
        return nil
    }

    u2 := &UploadRequest{
        cfr:       u.cfr.Next,
        ctrl:      u.ctrl,
        startTime: u.startTime,
    }

    return u2
}

type UploadAction struct {
    id  int32
    req *UploadRequest
}

func (a *UploadAction) ID() int32 {
    return a.id
}

func (a *UploadAction) Do(ctx context.Context) {
    startTime := domain.Now()

    bytes := pools.Bytes.GetLen(int(a.req.cfr.ChunkSize))
    defer pools.Bytes.Put(bytes)

    // Calculate offset based on chunk id and the chunk size
    offset := a.id * a.req.cfr.ChunkSize

    // We try to read the chunk, if it failed we try one more time
    n, err := a.req.file.ReadAt(bytes, int64(offset))
    if err != nil && err != io.EOF {
        logger.Warn("got error in ReadFile (Upload)", zap.Error(err))
        a.req.parts <- a.id
        return
    }

    // If we read 0 bytes then something is wrong
    if n == 0 {
        logger.Fatal("read zero bytes from file",
            zap.String("FilePath", a.req.cfr.FilePath),
            zap.Int32("TotalParts", a.req.cfr.TotalParts),
            zap.Int32("ChunkSize", a.req.cfr.ChunkSize),
        )
    }

    req := &msg.FileSavePart{
        TotalParts: a.req.cfr.TotalParts,
        Bytes:      bytes[:n],
        FileID:     a.req.cfr.FileID,
        PartID:     a.id + 1,
    }
    reqCB := request.NewCallback(
        0, 0, domain.NextRequestID(), msg.C_FileSavePart, req,
        func() {
            atomic.AddInt32(&a.req.failedActions, 1)
            a.req.parts <- a.id
        },
        func(res *rony.MessageEnvelope) {
            switch res.Constructor {
            case msg.C_Bool:
                a.req.addToUploaded(a.id)
                logger.Debug("upload action done",
                    zap.String("ID", a.req.GetID()),
                    zap.Int32("PartID", a.ID()),
                    zap.Duration("D", domain.Now().Sub(startTime)),
                )
            case rony.C_Error:
                x := &rony.Error{}
                _ = x.Unmarshal(res.Message)
                logger.Warn("received Error response (Upload)",
                    zap.Int32("PartID", a.id+1),
                    zap.String("Code", x.Code),
                    zap.String("Item", x.Items),
                )
                atomic.AddInt32(&a.req.failedActions, 1)
                a.req.parts <- a.id
            default:
                logger.Fatal("received unexpected response (Upload)", zap.String("C", registry.ConstructorName(res.Constructor)))
                return
            }
        },
        nil, false, 0, domain.HttpRequestTimeout,
    )

    a.req.ctrl.network.HttpCommand(nil, reqCB)
}
