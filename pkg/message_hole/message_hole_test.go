package messageHole

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/internal/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"testing"
)

func init() {
	repo.InitRepo("./_data", false)
}
func TestHole(t *testing.T) {
	peerID := domain.RandomInt63()
	// peerID := int64(239992)
	peerType := int32(1)

	// Test 1
	logs.Info("Test 1")
	peerID = domain.RandomInt63()
	InsertFill(0, peerID, peerType, 10, 11)
	InsertFill(0, peerID, peerType, 11, 13)
	InsertFill(0, peerID, peerType, 15, 16)
	InsertFill(0, peerID, peerType, 17, 19)
	logs.Info(PrintHole(0,peerID, peerType))

	// Test 2
	logs.Info("Test 2")
	peerID = domain.RandomInt63()
	InsertFill(0, peerID, peerType, 6, 8)
	InsertFill(0, peerID, peerType, 19, 20)
	InsertFill(0, peerID, peerType, 12, 12)
	InsertFill(0, peerID, peerType, 12, 12)
	InsertFill(0, peerID, peerType, 13, 14)
	InsertFill(0, peerID, peerType, 15, 15)
	logs.Info(PrintHole(0, peerID, peerType))

	// Test 3
	logs.Info("Test 3")
	peerID = domain.RandomInt63()
	InsertFill(0,peerID, peerType, 12, 12)
	InsertFill(0,peerID, peerType, 101, 120)
	InsertFill(0,peerID, peerType, 110, 120)
	InsertFill(0,peerID, peerType, 140, 141)
	InsertFill(0,peerID, peerType, 141, 142)
	InsertFill(0,peerID, peerType, 143, 143)
	logs.Info(PrintHole(0, peerID, peerType))

	// Test 4
	logs.Info("Test 4")
	peerID = domain.RandomInt63()
	InsertFill(0, peerID, peerType, 1001, 1001)
	InsertFill(0, peerID, peerType, 800, 900)
	InsertFill(0, peerID, peerType, 700, 850)
	InsertFill(0, peerID, peerType, 700, 799)
	InsertFill(0, peerID, peerType, 701, 799)
	InsertFill(0, peerID, peerType, 701, 801)
	InsertFill(0, peerID, peerType, 100, 699)
	logs.Info(PrintHole(0, peerID, peerType))

	// Test 5
	logs.Info("Test 5")
	peerID = domain.RandomInt64(0)
	InsertFill(0, peerID, peerType, 1001, 1001)
	InsertFill(0, peerID, peerType, 400, 500)
	InsertFill(0, peerID, peerType, 600, 700)
	InsertFill(0, peerID, peerType, 399, 699)
	logs.Info(PrintHole(0, peerID, peerType))
}
