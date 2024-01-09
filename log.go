package main

import (
	"fmt"
	"log"
)

func setupLog(name string) CloseFunc {
	logFile, closeFunc := openFile(fmt.Sprintf("organiser-filene-dine-%s.log", name))
	log.SetOutput(logFile)
	return closeFunc
}
