package logger

import (
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	l          = log.New(os.Stdout, "", log.LstdFlags)
	timeFormat = time.Now().Format("2006-01-02 15:04:05")
	logLevel   = INFO
)

func SetLogLevelFromString(level string) {
	switch level {
	case "DEBUG":
		logLevel = DEBUG
		break
	case "INFO":
		logLevel = INFO
		break
	case "WARN":
		logLevel = WARN
		break
	case "ERROR":
		logLevel = ERROR
		break
	case "FATAL":
		logLevel = FATAL
		break
	default:
		logLevel = INFO
		Info("log level not properly set, use default INFO")
	}
}

func logWithPrefix(prefix, msg string) {
	l.SetPrefix(prefix)
	l.Println(msg)
}

func Debug(msg string) {
	if logLevel <= DEBUG {
		logWithPrefix(timeFormat+" [DEBUG] ", msg)
	}
}

func Info(msg string) {
	if logLevel <= INFO {
		logWithPrefix(timeFormat+" [INFO] ", msg)
	}
}

func Warn(msg string) {
	if logLevel <= WARN {
		logWithPrefix(timeFormat+" [WARN] ", msg)
	}
}

func Error(msg string) {
	if logLevel <= ERROR {
		logWithPrefix(timeFormat+" [ERROR] ", msg)
	}
}

func Fatal(msg string) {
	if logLevel <= FATAL {
		logWithPrefix(timeFormat+" [FATAL] ", msg)
	}
}
