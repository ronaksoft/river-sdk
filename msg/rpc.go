package msg

/*
   Creation Time: 2018 - Apr - 07
   Created by:  Ehsan N. Moosa (ehsan)
   Maintainers:
       1.  Ehsan N. Moosa (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

func ErrorMessage(out *MessageEnvelope, errCode, errItem string) {
	errMessage := PoolError.Get().(*Error)
	defer PoolError.Put(errMessage)
	errMessage.Code = errCode
	errMessage.Items = errItem
	ResultError(out, errMessage)
	return
}
