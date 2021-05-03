package mini

import (
	"context"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
)

/*
   Creation Time: 2021 - May - 01
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *River) clientSendMessageMedia(in, out *rony.MessageEnvelope, da *DelegateAdapter) {
	reqMedia := &msg.ClientSendMessageMedia{}
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	// support IOS file path
	if strings.HasPrefix(reqMedia.FilePath, "file://") {
		reqMedia.FilePath = reqMedia.FilePath[7:]
	}
	if strings.HasPrefix(reqMedia.ThumbFilePath, "file://") {
		reqMedia.ThumbFilePath = reqMedia.ThumbFilePath[7:]
	}

	fileID := tools.SecureRandomInt63(0)
	thumbID := int64(0)
	reqMedia.FileUploadID = fmt.Sprintf("%d", fileID)
	reqMedia.FileID = fileID
	if reqMedia.ThumbFilePath != "" {
		thumbID = tools.SecureRandomInt63(0)
		reqMedia.ThumbID = thumbID
		reqMedia.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}

	checkSha256 := true
	switch reqMedia.MediaType {
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		for _, attr := range reqMedia.Attributes {
			if attr.Type == msg.DocumentAttributeType_AttributeTypeAudio {
				x := &msg.DocumentAttributeAudio{}
				_ = x.Unmarshal(attr.Data)
				if x.Voice {
					checkSha256 = false
				}
			}
		}
	default:
		panic("Invalid MediaInputType")
	}

	if thumbID != 0 {
		err := r.uploadFile(in, nil, thumbID, reqMedia.ThumbFilePath, reqMedia.Peer.ID)
		if err != nil {
			rony.ErrorMessage(out, in.RequestID, "E100", err.Error())
			da.OnComplete(out)
			return
		}
	}

	var (
		fileLocation *msg.FileLocation
	)
	if checkSha256 {
		fileLocation, _ = r.checkSha256(reqMedia)
	}

	// Create SendMessageMedia Request
	x := &msg.MessagesSendMedia{
		Peer:       reqMedia.Peer,
		ClearDraft: reqMedia.ClearDraft,
		RandomID:   tools.RandomInt64(0),
		ReplyTo:    reqMedia.ReplyTo,
	}

	if fileLocation != nil {
		// File already uploaded
		x.MediaType = msg.InputMediaType_InputMediaTypeDocument
		doc := &msg.InputMediaDocument{
			Caption:    reqMedia.Caption,
			Attributes: reqMedia.Attributes,
			Entities:   reqMedia.Entities,
			Document: &msg.InputDocument{
				ID:         fileLocation.FileID,
				AccessHash: fileLocation.AccessHash,
				ClusterID:  fileLocation.ClusterID,
			},
			TinyThumbnail: reqMedia.TinyThumb,
		}
		if thumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      thumbID,
				FileName:    "thumb_" + reqMedia.FileName,
				MD5Checksum: "",
			}
		}
		x.MediaData, _ = doc.Marshal()
	} else {
		err := r.uploadFile(in, da, fileID, reqMedia.FilePath, reqMedia.Peer.ID)
		if err != nil {
			rony.ErrorMessage(out, in.RequestID, "E100", err.Error())
			da.OnComplete(out)
			return
		}
		// File just uploaded
		x.MediaType = msg.InputMediaType_InputMediaTypeUploadedDocument
		doc := &msg.InputMediaUploadedDocument{
			MimeType:   reqMedia.FileMIME,
			Attributes: reqMedia.Attributes,
			Caption:    reqMedia.Caption,
			Entities:   reqMedia.Entities,
			File: &msg.InputFile{
				FileID:      fileID,
				FileName:    reqMedia.FileName,
				MD5Checksum: "",
			},
			TinyThumbnail: reqMedia.TinyThumb,
		}
		if thumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      thumbID,
				FileName:    "thumb_" + reqMedia.FileName,
				MD5Checksum: "",
			}
		}
		x.MediaData, _ = doc.Marshal()
	}

	reqBuff, _ := x.Marshal()
	r.networkCtrl.HttpCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(x.RandomID),
			Message:     reqBuff,
			Header:      domain.TeamHeader(domain.GetTeamID(in), domain.GetTeamAccess(in)),
		},
		da.OnTimeout, da.OnComplete,
	)

}
func (r *River) checkSha256(req *msg.ClientSendMessageMedia) (*msg.FileLocation, error) {
	h, _ := domain.CalculateSha256(req.FilePath)
	if len(h) == 0 {
		return nil, domain.ErrDoesNotExists
	}
	// Check File stats and return error if any problem exists
	var fileSize int32
	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		return nil, err
	} else {
		fileSize = int32(fileInfo.Size())
		if fileSize <= 0 {
			return nil, domain.ErrInvalidData
		} else if fileSize > fileCtrl.MaxFileSizeAllowedSize {
			return nil, domain.ErrFileTooLarge
		}
	}

	envelope := &rony.MessageEnvelope{}
	envelope.Fill(tools.RandomUint64(0), msg.C_FileGetBySha256, &msg.FileGetBySha256{
		Sha256:   h,
		FileSize: fileSize,
	})

	ctx, cancelFunc := context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
	defer cancelFunc()
	res, err := r.networkCtrl.SendHttp(ctx, envelope)
	if err != nil {
		return nil, err
	}
	switch res.Constructor {
	case msg.C_FileLocation:
		x := &msg.FileLocation{}
		_ = x.Unmarshal(res.Message)
		return x, nil
	case rony.C_Error:
		x := &rony.Error{}
		_ = x.Unmarshal(res.Message)
		return nil, x
	}
	return nil, domain.ErrServer
}
func (r *River) uploadFile(in *rony.MessageEnvelope, da *DelegateAdapter, fileID int64, filePath string, peerID int64) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var (
		fileSize int64
	)
	// Check File stats and return error if any problem exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	} else {
		fileSize = fileInfo.Size()
		if fileSize <= 0 {
			return domain.ErrInvalidData
		} else if fileSize > fileCtrl.MaxFileSizeAllowedSize {
			return domain.ErrFileTooLarge
		}
	}

	// Calculate number of parts based on our chunk size
	totalParts := int32(0)
	dividend := int32(fileSize / fileCtrl.DefaultChunkSize)
	if fileSize%int64(fileCtrl.DefaultChunkSize) > 0 {
		totalParts = dividend + 1
	} else {
		totalParts = dividend
	}

	for partIndex := int32(0); partIndex < totalParts; partIndex++ {
		logs.Info("SavePart",
			zap.Int32("PartID", partIndex), zap.Int32("Total", totalParts),
			zap.Int64("FileSize", fileSize),
		)
		err = r.savePart(in, f, fileID, partIndex, totalParts)
		if err != nil {
			logs.Warn("Error On SavePart (MiniSDK)", zap.Error(err))
			return err
		}

		if da != nil {
			da.OnProgress(int64(float64(partIndex) / float64(totalParts) * 100))
		}
	}

	return nil
}
func (r *River) savePart(in *rony.MessageEnvelope, f io.Reader, fileID int64, partIndex, totalParts int32) error {
	var buf [fileCtrl.DefaultChunkSize]byte
	n, err := f.Read(buf[:])
	if err != nil {
		return err
	}
	req := &msg.FileSavePart{
		FileID:     fileID,
		PartID:     partIndex + 1,
		TotalParts: totalParts,
		Bytes:      buf[:n],
	}
	reqBuf := pools.Buffer.FromProto(req)
	defer pools.Buffer.Put(reqBuf)
	r.networkCtrl.HttpCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_FileSavePart,
			RequestID:   tools.RandomUint64(0),
			Message:     *reqBuf.Bytes(),
			Header:      domain.TeamHeader(domain.GetTeamID(in), domain.GetTeamAccess(in)),
		},
		func() {
			err = domain.ErrRequestTimeout
		},
		func(m *rony.MessageEnvelope) {
			switch m.Constructor {
			case msg.C_Bool:
				err = nil
			case rony.C_Error:
				x := &rony.Error{}
				_ = x.Unmarshal(m.Message)
				err = x
			default:
				err = domain.ErrServer
			}
		},
	)
	return err
}
