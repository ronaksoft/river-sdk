package controller

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

var (
	sentmx          sync.Mutex
	sentPackets     []*msg.ProtoMessage
	recvmx          sync.Mutex
	receivedPackets []*msg.ProtoMessage
)

func init() {
	sentPackets = make([]*msg.ProtoMessage, 0)
	receivedPackets = make([]*msg.ProtoMessage, 0)
}

func logSentPacket(m *msg.ProtoMessage) {
	sentmx.Lock()
	sentPackets = append(sentPackets, m)
	sentmx.Unlock()
}

func logReceivedPacket(m *msg.ProtoMessage) {
	recvmx.Lock()
	receivedPackets = append(receivedPackets, m)
	recvmx.Unlock()
}

func GetLoggedReceivedPackets() []*msg.ProtoMessage {
	return receivedPackets
}
func GetLoggedSentPackets() []*msg.ProtoMessage {
	return sentPackets
}

func ClearLoggedPackets() {
	sentmx.Lock()
	sentPackets = sentPackets[:0]
	sentmx.Unlock()

	recvmx.Lock()
	receivedPackets = receivedPackets[:0]
	recvmx.Unlock()
}
