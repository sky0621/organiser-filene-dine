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

const (
	operationAll          = 9
	operationPrepare      = 1
	operationCreateOutDir = 2
	operationCopy         = 3
)

const errorListName = "errorList.txt"

func main() {
	cfg := getConfig()

	createDirectory(filepath.Join(cfg.ToDir, metaDir))

	/****************************************************************
	 * create copy list
	 */
	if cfg.Operation == operationPrepare || cfg.Operation == operationAll {
		listUp(cfg)
	}

	/****************************************************************
	 * create output directory
	 */
	if cfg.Operation == operationCreateOutDir || cfg.Operation == operationAll {
		createOutputDir(cfg)
	}

	/****************************************************************
	 * copy
	 */
	if cfg.Operation == operationCopy || cfg.Operation == operationAll {
		closeFunc := setupLog("copy")
		defer closeFunc()

		log.Printf("START: %s\n", time.Now().Format(time.RFC3339))

		copyListFile, err := os.Open(filepath.Join(cfg.ToDir, copyListFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := copyListFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		errorList, err := os.Create(filepath.Join(cfg.ToDir, errorListName))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := errorList.Close(); err != nil {
				log.Fatal(err)
			}
		}()

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

}

func copyFile(fromPath string, toPath string, errorList *os.File, semaphore chan struct{}, wg *sync.WaitGroup) error {
	defer func() {
		<-semaphore // 処理後にチャネルから値を抜き出さないと、次の goroutine が起動できない
	}()
	defer wg.Done()

	fromFile, err := os.Open(fromPath)
	if err != nil {
		log.Println("[[[ failed to open fromFile ]]]")
		log.Println(err)
		_, err = errorList.WriteString(fromPath + "\n")
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	defer func() {
		if err := fromFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	toFile, err := os.Create(toPath)
	if err != nil {
		log.Println("[[[ failed to create toFile ]]]")
		log.Println(err)
		_, err = errorList.WriteString(fmt.Sprintf("[from:%s] [to:%s]\n", fromPath, toPath))
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	defer func() {
		if err := toFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	if _, err := io.Copy(toFile, fromFile); err != nil {
		log.Println("[[[ failed to copy ]]]")
		log.Println(err)
		_, err = errorList.WriteString(fmt.Sprintf("[from:%s] [to:%s]\n", fromPath, toPath))
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	log.Printf("copied:[from:%s] [to:%s]\n", fromPath, toPath)

	return nil
}
