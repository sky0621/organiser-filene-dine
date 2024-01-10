package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

const metaDir = ".organiser-filene-dine"

const seps = "#-#-#$%&**&%$#-#-#"

const outputDirSetFileName = "outputDirSet.txt"

type CloseFunc func()

func open(path string) (*os.File, CloseFunc) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	return f, func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}
}

func openFile(path string) (*os.File, CloseFunc) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	return f, func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}
}

func renameFile(oldPath string, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func createDirectory(path string) {
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		if strings.Contains(err.Error(), "file exists") {
			return
		}
		log.Fatal(err)
	}
}

func getOutputDirSetFilePath(rootPath string) string {
	return filepath.Join(rootPath, metaDir, outputDirSetFileName)
}

func getExt(fileName string) string {
	return filepath.Ext(fileName)
}

func contains(strs []string, s string) bool {
	ls := strings.ToLower(s)
	for _, str := range strs {
		if str == ls {
			return true
		}
	}
	return false
}
