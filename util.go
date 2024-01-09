package main

import (
	"log"
	"os"
	"path/filepath"
)

type CloseFunc func()

func openFile(path string) (*os.File, CloseFunc) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return f, func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}
}

func getExt(fileName string) string {
	return filepath.Ext(fileName)
}
