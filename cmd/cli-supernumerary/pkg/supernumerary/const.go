package supernumerary

type TickerAction int

const (
	TickerActionNone             = 0
	TickerActionSendMessage      = 1
	TickerActionSendFile         = 2
	TickerActionSendGroupMessage = 3
)
