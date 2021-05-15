package call

type Callback struct {
	OnUpdate        func(constructor int32, b []byte)
	InitStream      func(audio, video bool) (err error)
	InitConnection  func(connId int32, b []byte) (id int64, err error)
	CloseConnection func(connId int32, all bool) (err error)
	GetOfferSDP     func(connId int32) (res []byte, err error)
	GetAnswerSDP    func(connId int32, req []byte) (res []byte, err error)
	SetAnswerSDP    func(connId int32, b []byte) (err error)
	AddIceCandidate func(connId int32, b []byte) (err error)
}
