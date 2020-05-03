package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
)

var (
	wg       *sync.WaitGroup
	testCase int
	test     *testing.T
)

func init() {
	logs.Info("Creating New River SDK Instance")
	r := new(River)
	conInfo := new(RiverConnection)
	r.SetConfig(&RiverConfig{
		DbPath: "./_data/",
		DbID:   "test",
		// ServerKeysFilePath:     "./keys.json",
		ServerEndpoint:         "ws://new.river.im",
		MainDelegate:           new(MainDelegateDummy),
		FileDelegate:           new(FileDelegateDummy),
		LogLevel:               int(zapcore.DebugLevel),
		DocumentAudioDirectory: "./_files/audio",
		DocumentVideoDirectory: "./_files/video",
		DocumentPhotoDirectory: "./_files/photo",
		DocumentFileDirectory:  "./_files/file",
		DocumentCacheDirectory: "./_files/cache",
		DocumentLogDirectory:   "./_files/logs",
		ConnInfo:               conInfo,
	})
	err := r.AppStart()
	if err != nil {
		panic(err)
	}
	_River = r
}

