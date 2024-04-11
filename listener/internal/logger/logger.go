package logger

import (
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota // implicitly set to 0
	INFO                  // implicitly set to 1
	WARN                  // implicitly set to 2
	ERROR                 // implicitly set to 3
	FATAL                 // implicitly set to 4
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
	case "INFO":
		logLevel = INFO
	case "WARN":
		logLevel = WARN
	case "ERROR":
		logLevel = ERROR
	case "FATAL":
		logLevel = FATAL
	default:
		logLevel = INFO
		Info("log level not properly set, using default INFO")
	}
}

func logWithPrefix(prefix, msg string) {
	currentTimeFormat := time.Now().Format("2006-01-02 15:04:05")
	l.SetPrefix(currentTimeFormat + prefix)
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
		os.Exit(1) // terminate the program with an error when Fatal is called
	}
}
