package main

import (
	"log"
)

func setupLog(path string) CloseFunc {
	logFile, closeFunc := openFile(path)
	log.SetOutput(logFile)
	return closeFunc
}
