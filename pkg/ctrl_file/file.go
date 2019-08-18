package fileCtrl

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"io/ioutil"
	"os"
	"time"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/


type Controller struct {
	network *networkCtrl.Controller
}

func New(network *networkCtrl.Controller) *Controller {
	ctrl := new(Controller)
	ctrl.network = network

	return ctrl
}


// Upload file to server
func (ctrl *Controller) Upload(fileID int64, req *msg.ClientPendingMessage) error {
	x := new(msg.ClientSendMessageMedia)
	err := x.Unmarshal(req.Media)
	if err != nil {
		return err
	}

	file, err := os.Open(x.FilePath)
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if x.FileName == "" {
		x.FileName = fileInfo.Name()
	}
	fileSize := fileInfo.Size() // size in Byte
	if fileSize > domain.FileMaxAllowedSize {
		return domain.ErrMaxFileSize
	}

	return nil
}

// Download add download request
func (ctrl *Controller) Download(userMessage *msg.UserMessage) {
	filesStatus, _ := repo.Files.GetStatus(userMessage.ID)
	if filesStatus != nil {

	}

	switch userMessage.MediaType {
	case msg.MediaTypeEmpty:
	case msg.MediaTypeDocument:
		x := new(msg.MediaDocument)
		_ = x.Unmarshal(userMessage.Media)

	default:
		return
	}


}

// DownloadAccountPhoto download account photo from server its sync
func (ctrl *Controller) DownloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) (string, error) {
	var filePath string
	err := ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := new(msg.FileGet)
		req.Location = new(msg.InputFileLocation)
		if isBig {
			req.Location.ClusterID = photo.PhotoBig.ClusterID
			req.Location.AccessHash = photo.PhotoBig.AccessHash
			req.Location.FileID = photo.PhotoBig.FileID
			req.Location.Version = 0
		} else {
			req.Location.ClusterID = photo.PhotoSmall.ClusterID
			req.Location.AccessHash = photo.PhotoSmall.AccessHash
			req.Location.FileID = photo.PhotoSmall.FileID
			req.Location.Version = 0
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = GetAccountAvatarPath(userID, req.Location.FileID)
		res, err := ctrl.network.SendHttp(envelop)
		if err != nil {
			return err
		}

		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return fmt.Errorf("received error response {userID: %d,  %s }", userID, strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return err
			}

			// save to DB
			return nil
		default:
			return fmt.Errorf("received unknown response constructor {UserId : %d}", userID)
		}

	})
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// DownloadGroupPhoto download group photo from server its sync
func (ctrl *Controller) DownloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) (string, error) {
	var filePath string
	err := ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := new(msg.FileGet)
		req.Location = new(msg.InputFileLocation)
		if isBig {
			req.Location.ClusterID = photo.PhotoBig.ClusterID
			req.Location.AccessHash = photo.PhotoBig.AccessHash
			req.Location.FileID = photo.PhotoBig.FileID
			req.Location.Version = 0
		} else {
			req.Location.ClusterID = photo.PhotoSmall.ClusterID
			req.Location.AccessHash = photo.PhotoSmall.AccessHash
			req.Location.FileID = photo.PhotoSmall.FileID
			req.Location.Version = 0
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = GetGroupAvatarPath(groupID, req.Location.FileID)
		res, err := ctrl.network.SendHttp(envelop)
		if err != nil {
			return err
		}
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return fmt.Errorf("received error response {GroupID: %d,  %s }", groupID, strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return err
			}

			// save to DB
			return nil

		default:
			return fmt.Errorf("received unknown response constructor {GroupID : %d}", groupID)
		}

	})
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// DownloadThumbnail download thumbnail from server its sync
func (ctrl *Controller) DownloadThumbnail(fileID int64, accessHash uint64, clusterID, version int32) (string, error) {
	filePath := ""
	err := ronak.Try(10, 100*time.Millisecond, func() error {
		req := new(msg.FileGet)
		req.Location = &msg.InputFileLocation{
			AccessHash: accessHash,
			ClusterID:  clusterID,
			FileID:     fileID,
			Version:    version,
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = GetThumbnailPath(fileID, clusterID)
		res, err := ctrl.network.SendHttp(envelop)
		if err != nil {
			return err
		}

		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return fmt.Errorf("received error response {%s}", strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return err
			}

			// save to DB
			return nil

		default:
			return nil
		}

	})
	if err != nil {
		return "", err
	}
	return filePath, nil
}

