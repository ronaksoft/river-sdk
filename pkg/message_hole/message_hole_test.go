package messageHole

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"testing"
)

func init() {
	repo.InitRepo("./_data", false)
}
func TestHole(t *testing.T) {
	peerID := ronak.RandomInt64(0)
	// peerID := int64(239992)
	peerType := int32(1)

	// Test 1
	logs.Info("Test 1")
	peerID = ronak.RandomInt64(0)
	InsertFill(peerID, peerType, 10, 11)
	InsertFill(peerID, peerType, 11, 13)
	InsertFill(peerID, peerType, 15, 16)
	InsertFill(peerID, peerType, 17, 19)
	logs.Info(PrintHole(peerID, peerType))

	// Test 2
	logs.Info("Test 2")
	peerID = ronak.RandomInt64(0)
	InsertFill(peerID, peerType, 6, 8)
	InsertFill(peerID, peerType, 19, 20)
	InsertFill(peerID, peerType, 12, 12)
	InsertFill(peerID, peerType, 12, 12)
	InsertFill(peerID, peerType, 13, 14)
	InsertFill(peerID, peerType, 15, 15)
	logs.Info(PrintHole(peerID, peerType))

	// Test 3
	logs.Info("Test 3")
	peerID = ronak.RandomInt64(0)
	InsertFill(peerID, peerType, 12, 12)
	InsertFill(peerID, peerType, 101, 120)
	InsertFill(peerID, peerType, 110, 120)
	InsertFill(peerID, peerType, 140, 141)
	InsertFill(peerID, peerType, 141, 142)
	InsertFill(peerID, peerType, 143, 143)
	logs.Info(PrintHole(peerID, peerType))

	// Test 4
	logs.Info("Test 4")
	peerID = ronak.RandomInt64(0)
	InsertFill(peerID, peerType, 1001, 1001)
	InsertFill(peerID, peerType, 800, 900)
	InsertFill(peerID, peerType, 700, 850)
	InsertFill(peerID, peerType, 700, 799)
	InsertFill(peerID, peerType, 701, 799)
	InsertFill(peerID, peerType, 701, 801)
	InsertFill(peerID, peerType, 100, 699)
	logs.Info(PrintHole(peerID, peerType))

	// Test 5
	logs.Info("Test 5")
	peerID = ronak.RandomInt64(0)
	InsertFill(peerID, peerType, 1001, 1001)
	InsertFill(peerID, peerType, 400, 500)
	InsertFill(peerID, peerType, 600, 700)
	InsertFill(peerID, peerType, 399, 699)
	logs.Info(PrintHole(peerID, peerType))
}
