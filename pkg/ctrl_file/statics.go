package fileCtrl

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"mime"
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
	dirAudio string
	dirFile  string
	dirPhoto string
	dirVideo string
	dirCache string
)

// SetRootFolders directory paths to Download files
func SetRootFolders(audioDir, fileDir, photoDir, videoDir, cacheDir string) {
	dirAudio = audioDir
	dirFile = fileDir
	dirPhoto = photoDir
	dirVideo = videoDir
	dirCache = cacheDir
}

func GetFilePath(clientFile *msg.ClientFile) string {
	switch clientFile.Type {
	case msg.ClientFileType_Message:
		return getMessageFilePath(clientFile.MimeType, clientFile.FileID)
	case msg.ClientFileType_AccountProfilePhoto:
		return getAccountProfilePath(clientFile.UserID, clientFile.FileID)
	case msg.ClientFileType_GroupProfilePhoto:
		return getGroupProfilePath(clientFile.GroupID, clientFile.GroupID)
	case msg.ClientFileType_Thumbnail:
		return getThumbnailPath(clientFile.FileID, clientFile.ClusterID)
	}
	return ""
}

func getMessageFilePath(mimeType string, docID int64) string {
	mimeType = strings.ToLower(mimeType)
	var ext string
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
		return path.Join(dirCache, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "video/"):
		return path.Join(dirVideo, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "audio/"):
		return path.Join(dirAudio, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "image/"):
		return path.Join(dirPhoto, fmt.Sprintf("%d%s", docID, ext))
	default:
		return path.Join(dirFile, fmt.Sprintf("%d%s", docID, ext))
	}
}

func getThumbnailPath(fileID int64, clusterID int32) string {
	return path.Join(dirCache, fmt.Sprintf("%d%d%s", fileID, clusterID, ".jpg"))
}

func getAccountProfilePath(userID int64, fileID int64) string {
	return path.Join(dirCache, fmt.Sprintf("u%d_%d%s", userID, fileID, ".jpg"))
}

func getGroupProfilePath(groupID int64, fileID int64) string {
	return path.Join(dirCache, fmt.Sprintf("g%d_%d%s", groupID, fileID, ".jpg"))
}
