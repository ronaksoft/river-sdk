package messageHole

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"testing"
	"time"
)

func init() {
	err := repo.InitRepo("./_data", false)
	if err != nil {
		logs.Fatal(err.Error())
	}
}
func TestHole(t *testing.T) {
	peerID := ronak.RandomInt64(0)
	peerType := int32(1)
	go InsertHole(peerID, peerType, 0, 10)
	time.Sleep(time.Millisecond)
	go SetUpperFilled(peerID, peerType, 15)
	// go SetUpperFilled(peerID, peerType, 11)
	time.Sleep(5 * time.Second)
	logs.Info(PrintHole(peerID, peerType))

}
