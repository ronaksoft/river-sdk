package log

import (
	"fmt"
	"math"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	callerSkipOffset = 2
)

var (
	_LOG      *zap.Logger
	LOG_LEVEL zap.AtomicLevel
	logger    func(logLevel int, msg string)
)

func init() {

	LOG_LEVEL = zap.NewAtomicLevelAt(zap.DebugLevel)
	logConfig := zap.NewProductionConfig()
	logConfig.Encoding = "console"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	logConfig.Level = LOG_LEVEL
	if l, err := logConfig.Build(); err != nil {
		os.Exit(1)
	} else {
		_LOG = l
	}

}

func SetLogLevel(l int) {
	if _LOG == nil {
		panic("logs is not initialized !!!")
	}
	LOG_LEVEL.SetLevel(zapcore.Level(l))
}

func SetLogger(fn func(logLevel int, msg string)) {
	logger = fn
}

func LOG_Debug(msg string, fields ...zap.Field) {
	if int(LOG_LEVEL.Level()) > int(zap.DebugLevel) {
		return
	}

	callerInfo := fnGetCallerInfo()

	if logger != nil {
		logger(int(zap.DebugLevel), msg+"\t"+callerInfo+"\t"+fnConvertFieldToString(fields...))
	} else {
		_LOG.Debug(msg, fields...)
	}
}

func LOG_Warn(msg string, fields ...zap.Field) {
	if int(LOG_LEVEL.Level()) > int(zap.WarnLevel) {
		return
	}
	callerInfo := fnGetCallerInfo()

	if logger != nil {
		logger(int(zap.WarnLevel), msg+"\t"+callerInfo+"\t"+fnConvertFieldToString(fields...))
	} else {
		_LOG.Warn(msg, fields...)
	}
}

func LOG_Info(msg string, fields ...zap.Field) {
	if int(LOG_LEVEL.Level()) > int(zap.InfoLevel) {
		return
	}
	callerInfo := fnGetCallerInfo()

	if logger != nil {
		logger(int(zap.InfoLevel), msg+"\t"+callerInfo+"\t"+fnConvertFieldToString(fields...))
	} else {
		_LOG.Info(msg, fields...)
	}
}

func LOG_Error(msg string, fields ...zap.Field) {
	if int(LOG_LEVEL.Level()) > int(zap.ErrorLevel) {
		return
	}
	callerInfo := fnGetCallerInfo()

	if logger != nil {
		logger(int(zap.ErrorLevel), msg+"\t"+callerInfo+"\t"+fnConvertFieldToString(fields...))
	} else {
		_LOG.Error(msg, fields...)
	}
}

func LOG_Fatal(msg string, fields ...zap.Field) {
	if int(LOG_LEVEL.Level()) > int(zap.FatalLevel) {
		return
	}
	callerInfo := fnGetCallerInfo()

	if logger != nil {
		logger(int(zap.FatalLevel), msg+"\t"+callerInfo+"\t"+fnConvertFieldToString(fields...))
	}
	//  The logger should calls os.Exit(1) when its FatalLevel
	_LOG.Fatal(msg, fields...)

}

func fnConvertFieldToString(fields ...zap.Field) string {
	sb := strings.Builder{}
	sb.WriteString("{")

	lastIndex := len(fields) - 1
	format := "%s : %v, "
	for idx, f := range fields {
		if idx == lastIndex {
			format = "%s : %v"
		}
		switch f.Type {

		case zapcore.ArrayMarshalerType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.(zapcore.ArrayMarshaler)))

		case zapcore.ObjectMarshalerType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.(zapcore.ObjectMarshaler)))

		case zapcore.BinaryType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.([]byte)))

		case zapcore.BoolType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Integer == 1))

		case zapcore.ByteStringType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.([]byte)))

		case zapcore.Complex128Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.(complex128)))

		case zapcore.Complex64Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.(complex64)))

		case zapcore.DurationType:
			sb.WriteString(fmt.Sprintf(format, f.Key, time.Duration(f.Integer)))

		case zapcore.Float64Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, math.Float64frombits(uint64(f.Integer))))

		case zapcore.Float32Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, math.Float32frombits(uint32(f.Integer))))

		case zapcore.Int64Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Integer))

		case zapcore.Int32Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, int32(f.Integer)))

		case zapcore.Int16Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, int16(f.Integer)))

		case zapcore.Int8Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, int8(f.Integer)))

		case zapcore.StringType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.String))

		case zapcore.TimeType:
			if f.Interface != nil {
				sb.WriteString(fmt.Sprintf(format, f.Key, time.Unix(0, f.Integer).In(f.Interface.(*time.Location))))

			} else {
				// Fall back to UTC if location is nil.
				sb.WriteString(fmt.Sprintf(format, f.Key, time.Unix(0, f.Integer)))

			}
		case zapcore.Uint64Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, uint64(f.Integer)))

		case zapcore.Uint32Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, uint32(f.Integer)))

		case zapcore.Uint16Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, uint16(f.Integer)))

		case zapcore.Uint8Type:
			sb.WriteString(fmt.Sprintf(format, f.Key, uint8(f.Integer)))

		case zapcore.UintptrType:
			sb.WriteString(fmt.Sprintf(format, f.Key, uintptr(f.Integer)))

		case zapcore.ReflectType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface))

		case zapcore.NamespaceType:
			//enc.OpenNamespace(f.Key)
			//counter++
		case zapcore.StringerType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.(fmt.Stringer).String()))

		case zapcore.ErrorType:
			sb.WriteString(fmt.Sprintf(format, f.Key, f.Interface.(error)))

		case zapcore.SkipType:
			break
		default:
			panic(fmt.Sprintf("unknown field type: %v", f))

		}

	}
	sb.WriteString("}")
	if lastIndex >= 0 {
		return sb.String()
	}
	return ""

}

func fnGetCallerInfo() string {

	// get line of code that called this
	callerInfo := ""

	_, fileName, line, ok := runtime.Caller(callerSkipOffset)
	if ok {
		fileName = path.Base(fileName)
		callerInfo = fmt.Sprintf("%s:%d", fileName, line)
	}

	return callerInfo
}
