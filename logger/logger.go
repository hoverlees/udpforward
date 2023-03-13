package logger

import (
	"log"
)

const LOG_LEVEL_DEBUG = 0
const LOG_LEVEL_INFO = 1
const LOG_LEVEL_WARN = 2
const LOG_LEVEL_ERROR = 3

var logLevel int = LOG_LEVEL_DEBUG

func SetLogLevel(level int) {
	logLevel = level
}

func SetLogLevelByName(levelName string) {
	Info("Set logger level to %s", levelName)
	if levelName == "INFO" {
		SetLogLevel(LOG_LEVEL_INFO)
	} else if levelName == "WARN" {
		SetLogLevel(LOG_LEVEL_WARN)
	} else if levelName == "ERROR" {
		SetLogLevel(LOG_LEVEL_ERROR)
	} else {
		SetLogLevel(LOG_LEVEL_DEBUG)
	}
}

func Debug(fmt string, v ...interface{}) {
	if logLevel > LOG_LEVEL_DEBUG {
		return
	}
	log.Printf("[DEBUG] "+fmt, v...)
}

func Info(fmt string, v ...interface{}) {
	if logLevel > LOG_LEVEL_INFO {
		return
	}
	log.Printf("[INFO] "+fmt, v...)
}

func Warn(fmt string, v ...interface{}) {
	if logLevel > LOG_LEVEL_WARN {
		return
	}
	log.Printf("[WARN] "+fmt, v...)
}

func Error(fmt string, v ...interface{}) {
	log.Printf("[ERROR] "+fmt, v...)
}

func Panic(fmt string, v ...interface{}) {
	log.Panicf("[Panic] "+fmt, v...)
}

func IsDebug() bool {
	return logLevel == LOG_LEVEL_DEBUG
}
