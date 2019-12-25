package controller

import (
	"sync"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
)

var (
	sentmx          sync.Mutex
	sentPackets     []*msg.ProtoMessage
	recvmx          sync.Mutex
	receivedPackets []*msg.ProtoMessage
	isEnabled       bool
)

func init() {
	isEnabled = true
	sentPackets = make([]*msg.ProtoMessage, 0)
	receivedPackets = make([]*msg.ProtoMessage, 0)
}

func logSentPacket(m *msg.ProtoMessage) {
	if isEnabled {
		sentmx.Lock()
		sentPackets = append(sentPackets, m)
		sentmx.Unlock()
	}
}

func logReceivedPacket(m *msg.ProtoMessage) {
	if isEnabled {
		recvmx.Lock()
		receivedPackets = append(receivedPackets, m)
		recvmx.Unlock()
	}
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

func StopLogginPackets() {
	isEnabled = false
	ClearLoggedPackets()
}
