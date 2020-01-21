package fileCtrl

import (
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
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

func minChunkSize(fileSize int64) int32 {
	minChunkSize := (fileSize / maxParts) >> 10
	dataRate := mon.GetDataTransferRate()
	if dataRate == 0 {
		dataRate = chunkSizesKB[len(chunkSizesKB)-1]
	}
	min := int32(math.Min(float64(minChunkSize), float64(dataRate)))
	for _, cs := range chunkSizesKB {
		if min > cs {
			continue
		}
		return cs << 10
	}
	return chunkSizesKB[len(chunkSizesKB)-1] << 10
}
