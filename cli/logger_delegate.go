package main

import "go.uber.org/zap"

type Logger struct{}

func (l *Logger) Log(logLevel int, msg string) {

	switch logLevel {
	case int(zap.DebugLevel):
		_Shell.Println("Debug LOG :", msg)
	case int(zap.WarnLevel):
		_Shell.Println(_Yellow("Warning LOG : %s", msg))
	case int(zap.InfoLevel):
		_Shell.Println(_MAGNETA("Info LOG : %s", msg))
	case int(zap.ErrorLevel):
		_Shell.Println(_RED("Error LOG : %s", msg))
	case int(zap.FatalLevel):
		_Shell.Println(_GREEN("Fatal LOG : %s", msg))
	}

}
