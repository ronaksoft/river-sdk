package shared

import (
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
)

// SuccessCallback timeout
type SuccessCallback func(response *msg.MessageEnvelope, elapsed time.Duration)

// TimeoutCallback function
type TimeoutCallback func(requestID uint64, elapsed time.Duration)

// UpdateApplier function
type UpdateApplier func(act Actor, u *msg.UpdateEnvelope)
