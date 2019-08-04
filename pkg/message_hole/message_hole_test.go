package messageHole

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"testing"
)

func init() {
	repo.InitRepo("./_data", false)
}
func TestHole(t *testing.T) {
	// peerID := ronak.RandomInt64(0)
	peerID := int64(234)
	peerType := int32(1)


	// InsertFill(peerID, peerType, 10, 10)
	InsertFill(peerID, peerType, 10, 11)

	// InsertFill(peerID, peerType, 12, 13)
	// InsertHole(peerID, peerType, 0, 100)
	// InsertFill(peerID, peerType, 101, 120)
	// InsertFill(peerID, peerType, 140, 141)
	logs.Info(PrintHole(peerID, peerType))

}
