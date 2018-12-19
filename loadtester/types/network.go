package types

import "git.ronaksoftware.com/ronak/riversdk/msg"

type Network interface {
	Send(msgEnvelope *msg.MessageEnvelope) error
}
