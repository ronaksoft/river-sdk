package fileCtrl

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"time"
)

/*
   Creation Time: 2019 - Sep - 07
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/


var (
	downloads map[int64]*downloadStatus
)

func init() {
	downloads = make(map[int64]*downloadStatus)
}

func persistDownloads() {
	saveSnapshot.EnterWithResult(nil, nil)
}

var saveSnapshot = ronak.NewFlusher(100, 1, time.Millisecond * 100, func(items []ronak.FlusherEntry) {
	if dBytes, err := json.Marshal(downloads); err == nil {
		_ = repo.System.SaveBytes("Downloads", dBytes)
	}
	for idx := range items {
		items[idx].Callback(nil)
	}
})