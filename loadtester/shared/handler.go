package shared

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// SuccessCallback timeout
type SuccessCallback func(response *msg.MessageEnvelope, elapsed time.Duration)

// TimeoutCallback function
type TimeoutCallback func(requestID uint64, elapsed time.Duration)
