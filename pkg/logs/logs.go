package logs

import (
	"fmt"
	"math"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	callerSkipOffset = 2
)

var (
	logLevel                 zap.AtomicLevel
	logger                   func(logLevel int, msg string)
	green, red, yellow, blue func(format string, a ...interface{}) string
	logFilePath              string
	logWriteToFile           bool
)

func init() {
	logLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	green = color.New(color.FgHiGreen).SprintfFunc()
	red = color.New(color.FgHiRed).SprintfFunc()
	yellow = color.New(color.FgHiYellow).SprintfFunc()
	blue = color.New(color.FgHiBlue).SprintfFunc()

}

func SetLogLevel(l int) {
	logLevel.SetLevel(zapcore.Level(l))
}

func SetLogger(fn func(logLevel int, msg string)) {
	logger = fn
}

func SetLogFilePath(filePath string) {
	logFilePath = filePath
	logWriteToFile = filePath != ""
}

func Message(msg string, fields ...zap.Field) {
	callerInfo := getCallerInfo()
	if logger != nil {
		logger(int(-99), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	} else {
		log(int(-99), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}
	if logWriteToFile {
		saveLog(int(-99), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}
}

func Debug(msg string, fields ...zap.Field) {
	if int(logLevel.Level()) > int(zap.DebugLevel) {
		return
	}

	callerInfo := getCallerInfo()

	if logger != nil {
		logger(int(zap.DebugLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	} else {
		log(int(zap.DebugLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}

	if logWriteToFile {
		saveLog(int(zap.DebugLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}
}

func Warn(msg string, fields ...zap.Field) {
	if int(logLevel.Level()) > int(zap.WarnLevel) {
		return
	}
	callerInfo := getCallerInfo()

	if logger != nil {
		logger(int(zap.WarnLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	} else {
		log(int(zap.WarnLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}

	if logWriteToFile {
		saveLog(int(zap.WarnLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}
}

func Info(msg string, fields ...zap.Field) {
	if int(logLevel.Level()) > int(zap.InfoLevel) {
		return
	}
	callerInfo := getCallerInfo()

	if logger != nil {
		logger(int(zap.InfoLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	} else {
		log(int(zap.InfoLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}

	if logWriteToFile {
		saveLog(int(zap.InfoLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}
}

func Error(msg string, fields ...zap.Field) {
	if int(logLevel.Level()) > int(zap.ErrorLevel) {
		return
	}
	callerInfo := getCallerInfo()

	if logger != nil {
		logger(int(zap.ErrorLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	} else {
		log(int(zap.ErrorLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}

	if logWriteToFile {
		saveLog(int(zap.ErrorLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}
}

func Fatal(msg string, fields ...zap.Field) {
	defer os.Exit(1)

	if int(logLevel.Level()) > int(zap.FatalLevel) {
		return
	}
	callerInfo := getCallerInfo()

	if logger != nil {
		logger(int(zap.FatalLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	} else {
		log(int(zap.FatalLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}

	if logWriteToFile {
		saveLog(int(zap.FatalLevel), callerInfo+"\t\t"+msg+"\t"+convertFieldToString(fields...))
	}

}

func convertFieldToString(fields ...zap.Field) string {
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

func getCallerInfo() string {

	// get line of code that called this
	callerInfo := ""

	_, fileName, line, ok := runtime.Caller(callerSkipOffset)
	if ok {
		fileName = path.Base(fileName)
		callerInfo = fmt.Sprintf("%s:%d", fileName, line)
	}

	return callerInfo
}

func log(logLevel int, msg string) {

	switch logLevel {
	case int(zap.DebugLevel):
		fmt.Println("DBG : \t", msg)
	case int(zap.WarnLevel):
		fmt.Println(yellow("WRN : \t %s", msg))
	case int(zap.InfoLevel):
		fmt.Println(green("INF : \t %s", msg))
	case int(zap.ErrorLevel):
		fmt.Println(red("ERR : \t %s", msg))
	case int(zap.FatalLevel):
		fmt.Println(red("FTL : \t %s", msg))
	default:
		fmt.Println(blue("MSG : \t %s", msg))
	}
}

func saveLog(logLevel int, msg string) {

	switch logLevel {
	case int(zap.DebugLevel):
		msg = "DBG : \t " + msg
	case int(zap.WarnLevel):
		msg = "WRN : \t " + msg
	case int(zap.InfoLevel):
		msg = "INF : \t " + msg
	case int(zap.ErrorLevel):
		msg = "ERR : \t " + msg
	case int(zap.FatalLevel):
		msg = "FTL : \t " + msg
	default:
		msg = "MSG : \t " + msg
	}

	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		f.WriteString(msg + "\n")
		f.Close()
	} else {
		fmt.Println(err)
	}
}
