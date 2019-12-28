package fileCtrl

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"math"
	"mime"
	"os"
	"path"
	"strings"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var (
	DirAudio string
	DirFile  string
	DirPhoto string
	DirVideo string
	DirCache string
)

// SetRootFolders directory paths to Download files
func SetRootFolders(audioDir, fileDir, photoDir, videoDir, cacheDir string) {
	DirAudio = audioDir
	_ = os.MkdirAll(audioDir, os.ModePerm)
	DirFile = fileDir
	_ = os.MkdirAll(fileDir, os.ModePerm)
	DirPhoto = photoDir
	_ = os.MkdirAll(photoDir, os.ModePerm)
	DirVideo = videoDir
	_ = os.MkdirAll(videoDir, os.ModePerm)
	DirCache = cacheDir
	_ = os.MkdirAll(cacheDir, os.ModePerm)
}

func GetFilePath(clientFile *msg.ClientFile) string {
	switch clientFile.Type {
	case msg.ClientFileType_Message:
		return getMessageFilePath(clientFile.MimeType, clientFile.FileID, clientFile.Extension)
	case msg.ClientFileType_AccountProfilePhoto:
		return getAccountProfilePath(clientFile.UserID, clientFile.FileID)
	case msg.ClientFileType_GroupProfilePhoto:
		return getGroupProfilePath(clientFile.GroupID, clientFile.FileID)
	case msg.ClientFileType_Thumbnail:
		return getThumbnailPath(clientFile.FileID, clientFile.ClusterID)
	}
	return ""
}

func getMessageFilePath(mimeType string, docID int64, ext string) string {
	mimeType = strings.ToLower(mimeType)
	if ext == "" {
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[len(exts)-1]
		}
	}

	// if the file is opus type,
	// means its voice file so it should be saved in cache folder
	// so user could not access to it by file manager
	switch {
	case mimeType == "audio/ogg":
		ext = ".ogg"
		return path.Join(DirCache, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "video/"):
		return path.Join(DirVideo, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "audio/"):
		return path.Join(DirAudio, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "image/"):
		return path.Join(DirPhoto, fmt.Sprintf("%d%s", docID, ext))
	default:
		return path.Join(DirFile, fmt.Sprintf("%d%s", docID, ext))
	}
}

func getThumbnailPath(fileID int64, clusterID int32) string {
	return path.Join(DirCache, fmt.Sprintf("%d%d%s", fileID, clusterID, ".jpg"))
}

func getAccountProfilePath(userID int64, fileID int64) string {
	return path.Join(DirCache, fmt.Sprintf("u%d_%d%s", userID, fileID, ".jpg"))
}

func getGroupProfilePath(groupID int64, fileID int64) string {
	return path.Join(DirCache, fmt.Sprintf("g%d_%d%s", groupID, fileID, ".jpg"))
}

func unique(intSlice []int32) []int32 {
	keys := make(map[int32]bool)
	list := make([]int32, 0)
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func bestChunkSize(fileSize int64) int32 {
	if fileSize <= maxChunkSize {
		return defaultChunkSize
	}
	minChunkSize := (fileSize / maxParts) >> 10
	dataRate := mon.GetDataTransferRate()
	max := int32(math.Max(float64(minChunkSize), float64(dataRate)))
	for _, cs := range chunkSizesKB {
		if max > cs {
			continue
		}
		return cs << 10
	}
	return chunkSizesKB[len(chunkSizesKB)-1] << 10
}

func minChunkSize(fileSize int64) int32 {
	minChunkSize := (fileSize / maxParts) >> 10
	dataRate := mon.GetDataTransferRate()
	min := int32(math.Min(float64(minChunkSize), float64(dataRate)))
	for _, cs := range chunkSizesKB {
		if min > cs {
			continue
		}
		return cs << 10
	}
	return chunkSizesKB[len(chunkSizesKB)-1] << 10
}
