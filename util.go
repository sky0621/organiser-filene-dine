package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func getCopyListFilePath(rootPath string) string {
	return filepath.Join(rootPath, metaDir, copyListFileName)
}

func getCopyListBackupFilePath(rootPath string) string {
	return filepath.Join(rootPath, metaDir, copyListFileName+"_"+time.Now().Format("20060102150405"))
}

func getOutputDirSetFilePath(rootPath string) string {
	return filepath.Join(rootPath, metaDir, outputDirSetFileName)
}

func getExt(fileName string) string {
	return strings.ToLower(filepath.Ext(fileName))
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

type oldNewPath struct {
	oldPath string
	newPath string
}

func readSomeBytes(path string) ([]byte, error) {
	f1, c1 := open(path)
	defer c1()

	byteArray := make([]byte, 1024)
	_, err := f1.Read(byteArray)
	if err != nil {
		if err.Error() == "EOF" {
			return nil, nil
		}
		return nil, err
	}

	return byteArray, nil
}

func containsSomeBytes(someBytes []byte, bytesFilePathMap map[string][]oldNewPath) bool {
	for k, _ := range bytesFilePathMap {
		if string(someBytes) == k {
			return true
		}
	}
	return false
}
