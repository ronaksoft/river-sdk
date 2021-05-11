package call

type Callback struct {
	OnUpdate        func(constructor int64, b []byte)
	CloseConnection func(connId int32)
	GetAnswerSDP    func(connId int32) []byte
	GetOfferSDP     func(connId int32) []byte
}
