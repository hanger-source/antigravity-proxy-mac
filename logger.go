package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logger *log.Logger
var logFile *os.File

func initLogger() {
	dir := ConfigDir()
	logPath := filepath.Join(dir, "antigravity-proxy.log")

	// Rotate if too large (>5MB)
	if info, err := os.Stat(logPath); err == nil && info.Size() > 5*1024*1024 {
		rotated := logPath + "." + time.Now().Format("20060102-150405")
		os.Rename(logPath, rotated)
	}

	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logger = log.New(os.Stderr, "", 0)
		return
	}
	logger = log.New(logFile, "", 0)

	logInfo("========================================")
	logInfo("antigravity-proxy-mac starting")
	logInfo("log file: %s", logPath)
	logInfo("config dir: %s", dir)
	logInfo("========================================")
}

func logInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("%s [INFO]  %s", time.Now().Format("2006-01-02 15:04:05.000"), msg)
	fmt.Fprintln(os.Stderr, entry)
	if logger != nil {
		logger.Println(entry)
	}
}

func logError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("%s [ERROR] %s", time.Now().Format("2006-01-02 15:04:05.000"), msg)
	fmt.Fprintln(os.Stderr, entry)
	if logger != nil {
		logger.Println(entry)
	}
}

func logWarn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("%s [WARN]  %s", time.Now().Format("2006-01-02 15:04:05.000"), msg)
	fmt.Fprintln(os.Stderr, entry)
	if logger != nil {
		logger.Println(entry)
	}
}

func logDebug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("%s [DEBUG] %s", time.Now().Format("2006-01-02 15:04:05.000"), msg)
	fmt.Fprintln(os.Stderr, entry)
	if logger != nil {
		logger.Println(entry)
	}
}
