package messageHole

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"testing"
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

	InsertHole(peerID, peerType, 0, 22551)
	InsertFill(peerID, peerType, 22552, 2250)
	InsertFill(peerID, peerType, 24000, 24001)
	// InsertHole(peerID, peerType, 0, 100)
	// InsertFill(peerID, peerType, 101, 120)
	// InsertFill(peerID, peerType, 140, 141)
	logs.Info(PrintHole(peerID, peerType))

}
