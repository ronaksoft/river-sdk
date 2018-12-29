package shared

import "git.ronaksoftware.com/ronak/riversdk/msg"

type Networker interface {
	Send(msgEnvelope *msg.MessageEnvelope) error
}
