package domain

import (
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
)

/*
   Creation Time: 2020 - Sep - 27
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/




func RecoverPanic(funcName string, extraInfo interface{}, compensationFunc func()) {
	if r := recover(); r != nil {
		logs.Error("Panic Recovered",
			zap.String("Func", funcName),
			zap.Any("Info", extraInfo),
			zap.Any("Recover", r),
		)
		if compensationFunc != nil {
			go compensationFunc()
		}
	}
}