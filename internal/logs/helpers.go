package logs

import (
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/*
   Creation Time: 2019 - Oct - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("06-01-02T15:04:05"))
}

func CleanUP() {
	lifeTime := 7 * 24 * time.Hour
	_ = filepath.Walk(_LogDir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".log") {
			if time.Since(info.ModTime()).Truncate(lifeTime) > 0 {
				_ = os.Remove(path)
			}
		}
		return nil
	})
}
