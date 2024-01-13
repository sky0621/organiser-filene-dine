package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const errorListName = "errorList.txt"
const execCopyLogFileName = "execCopy.log"

func execCopy(toDir string) {
	closeLogFile := openExecCopyLogFile(toDir)
	defer closeLogFile()

	log.Printf("START: %s\n", time.Now().Format(time.RFC3339))

	copyListFile, closeCopyListFile := open(getCopyListFilePath(toDir))
	defer closeCopyListFile()

	errorList, closeErrorListFile := openErrorListFile(toDir)
	defer closeErrorListFile()

	cpuNum := runtime.NumCPU()
	log.Printf("NumCPU: %d\n", cpuNum)

	// ★ 同時実行 goroutine 数の制御のためにチャネル用意
	semaphore := make(chan struct{}, cpuNum*6)

	wg := &sync.WaitGroup{}

	copyListFileScanner := bufio.NewScanner(copyListFile)
	for copyListFileScanner.Scan() {
		semaphore <- struct{}{}
		wg.Add(1)

		line := copyListFileScanner.Text()
		fromTo := strings.Split(line, seps)
		go func() {
			err := copyFile(fromTo[0], fromTo[1], errorList, semaphore, wg)
			if err != nil {
				log.Println(err)
			}
		}()
	}

	wg.Wait()
	log.Printf("END  : %s\n", time.Now().Format(time.RFC3339))
}

func copyFile(fromPath string, toPath string, errorList *os.File, semaphore chan struct{}, wg *sync.WaitGroup) error {
	defer func() {
		<-semaphore // 処理後にチャネルから値を抜き出さないと、次の goroutine が起動できない
	}()
	defer wg.Done()

	fromFile, err := os.Open(fromPath)
	if err != nil {
		log.Println("[[[ failed to open fromFile ]]]", err)
		writeErrorList(errorList, fromPath, toPath)
		return nil
	}
	defer func() {
		if err := fromFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	toFile, err := os.Create(toPath)
	if err != nil {
		log.Println("[[[ failed to create toFile ]]]", err)
		writeErrorList(errorList, fromPath, toPath)
		return nil
	}
	defer func() {
		if err := toFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	if _, err := io.Copy(toFile, fromFile); err != nil {
		log.Println("[[[ failed to copy ]]]", err)
		writeErrorList(errorList, fromPath, toPath)
		return nil
	}
	log.Printf("copied:[from:%s] [to:%s]\n", fromPath, toPath)

	return nil
}

func openExecCopyLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, execCopyLogFileName))
}

func openErrorListFile(rootPath string) (*os.File, CloseFunc) {
	return openFile(filepath.Join(rootPath, metaDir, errorListName))
}

func writeErrorList(errorList *os.File, fromPath string, toPath string) {
	_, err := errorList.WriteString(fmt.Sprintf("%s%s%s\n", fromPath, seps, toPath))
	if err != nil {
		log.Println(err)
	}
}
