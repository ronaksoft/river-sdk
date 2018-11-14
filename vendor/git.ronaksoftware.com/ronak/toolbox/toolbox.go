package ronak

import (
    "go.uber.org/zap"
    "os"
    "go.uber.org/zap/zapcore"
)

/*
    Creation Time: 2018 - Apr - 07
    Created by:  Ehsan N. Moosa (ehsan)
    Maintainers:
        1.  Ehsan N. Moosa (ehsan)
    Auditor: Ehsan N. Moosa
    Copyright Ronak Software Group 2018
*/

var (
    _LOG       *zap.Logger
    _LOG_LEVEL zap.AtomicLevel
)

type (
    M map[string]interface{}
    MS map[string]string
    MI map[string]int64
)

func init() {
    logConfig := zap.NewProductionConfig()
    logConfig.Encoding = "console"
    _LOG_LEVEL = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    logConfig.Level = _LOG_LEVEL
    if v, err := logConfig.Build(); err != nil {
        os.Exit(1)
    } else {
        _LOG = v
    }
}

func SetLogLevel(level int) {
    _LOG_LEVEL.SetLevel(zapcore.Level(level))
}
