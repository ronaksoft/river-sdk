package repo

import (
	"os"
)

/*
   Creation Time: 2021 - May - 03
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
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
