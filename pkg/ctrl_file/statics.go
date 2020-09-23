package fileCtrl

import (
	"fmt"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"math"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func bestChunkSize(fileSize int64) int32 {
	if fileSize <= maxChunkSize {
		return defaultChunkSize
	}
	minChunkSize := (fileSize / maxParts) >> 10
	dataRate := mon.GetDataTransferRate()
	if dataRate == 0 {
		dataRate = chunkSizesKB[len(chunkSizesKB)-1]
	}
	max := int32(math.Max(float64(minChunkSize), float64(dataRate)))
	for _, cs := range chunkSizesKB {
		if max > cs {
			continue
		}
		return cs << 10
	}
	return chunkSizesKB[len(chunkSizesKB)-1] << 10
}

func getRequestID(clusterID int32, fileID int64, accessHash uint64) string {
	return fmt.Sprintf("%d.%d.%d", clusterID, fileID, accessHash)
}
