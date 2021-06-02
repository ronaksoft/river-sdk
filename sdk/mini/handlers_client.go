package mini

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/minirepo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
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

func (r *River) clientSendMessageMedia(da request.Callback) {
	reqMedia := &msg.ClientSendMessageMedia{}
	if err := da.RequestData(reqMedia); err != nil {
		return
	}

	// support IOS file path
	reqMedia.FilePath = strings.TrimPrefix(reqMedia.FilePath, "file://")
	reqMedia.ThumbFilePath = strings.TrimPrefix(reqMedia.ThumbFilePath, "file://")

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
		err := r.uploadFile(da, thumbID, reqMedia.ThumbFilePath, false)
		if err != nil {
			da.Response(rony.C_Error, errors.New("00", err.Error()))
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
		da.OnProgress(100)
	} else {
		err := r.uploadFile(da, fileID, reqMedia.FilePath, true)
		if err != nil {
			da.Response(rony.C_Error, errors.New("00", err.Error()))
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

	da.Envelope().Fill(da.RequestID(), msg.C_MessagesSendMedia, x)
	r.network.HttpCommand(nil, da)
}
func (r *River) checkSha256(req *msg.ClientSendMessageMedia) (fl *msg.FileLocation, err error) {
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

	reqCB := request.NewCallback(
		0, 0, domain.NextRequestID(), msg.C_FileGetBySha256,
		&msg.FileGetBySha256{
			Sha256:   h,
			FileSize: fileSize,
		},
		func() {
			err = domain.ErrRequestTimeout
		},
		func(res *rony.MessageEnvelope) {
			switch res.Constructor {
			case msg.C_FileLocation:
				fl = &msg.FileLocation{}
				_ = fl.Unmarshal(res.Message)
				return
			case rony.C_Error:
				x := &rony.Error{}
				_ = x.Unmarshal(res.Message)
				err = x
			default:
				err = domain.ErrServer
			}
		},
		nil, false, 0, domain.HttpRequestTimeout,
	)
	r.network.HttpCommand(nil, reqCB)

	return
}
func (r *River) uploadFile(da request.Callback, fileID int64, filePath string, progress bool) error {
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
		logger.Info("SavePart",
			zap.Int32("PartID", partIndex), zap.Int32("Total", totalParts),
			zap.Int64("FileSize", fileSize),
		)
		err = r.savePart(da, f, fileID, partIndex, totalParts)
		if err != nil {
			logger.Warn("Error On SavePart (MiniSDK)", zap.Error(err))
			return err
		}

		if progress {
			da.OnProgress(int64(float64(partIndex) / float64(totalParts) * 100))
		}
	}

	return nil
}
func (r *River) savePart(da request.Callback, f io.Reader, fileID int64, partIndex, totalParts int32) error {
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
	reqCB := request.NewCallbackFromBytes(
		da.TeamID(), da.TeamAccess(), domain.NextRequestID(), msg.C_FileSavePart, *reqBuf.Bytes(),
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
		nil,
		false, 0, domain.HttpRequestTimeout,
	)
	r.network.HttpCommand(nil, reqCB)
	return err
}

func (r *River) clientGlobalSearch(da request.Callback) {
	req := &msg.ClientGlobalSearch{}
	if err := da.RequestData(req); err != nil {
		return
	}
	logger.Debug("local handler for ClientGlobalSearch called",
		zap.Int64("TeamID", da.TeamID()),
		zap.Any("Text", req.Text),
	)
	res := &msg.ClientSearchResult{}
	uniqueUsers := domain.MInt64B{}
	uniqueGroups := domain.MInt64B{}
	cUsers := minirepo.Users.SearchContacts(da.TeamID(), strings.ToLower(req.Text), int(req.Limit))
	for _, cu := range cUsers {
		uniqueUsers[cu.ID] = true
	}

	cGroups := minirepo.Groups.Search(da.TeamID(), strings.ToLower(req.Text), int(req.Limit))
	for _, g := range cGroups {
		uniqueGroups[g.ID] = true
	}

	res.Users, _ = minirepo.Users.ReadMany(uniqueUsers.ToArray()...)
	res.Groups, _ = minirepo.Groups.ReadMany(da.TeamID(), uniqueGroups.ToArray()...)
	res.MatchedUsers = append(res.MatchedUsers, res.Users...)
	res.MatchedGroups = append(res.MatchedGroups, res.Groups...)

	da.Response(msg.C_ClientSearchResult, res)
}
