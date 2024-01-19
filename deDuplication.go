package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const deDuplicationLogFileName = "deDuplication.log"

func deDuplication(toDir string) {
	closeLogFile := openDeDuplicationLogFile(toDir)
	defer closeLogFile()

	log.Printf("START: %s\n", time.Now().Format(time.RFC3339))
	bytesFilePathMap := make(map[string][]oldNewPath)
	if err := filepath.WalkDir(toDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println("failed to WalkDir", err)
			return err
		}

		fi, err := d.Info()
		if err != nil {
			log.Println("failed to get directory info", err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		_, fileName := filepath.Split(path)
		if fileName == ".DS_Store" {
			return nil
		}

		someBytes, err := readSomeBytes(path)
		if err != nil {
			return err
		}
		if someBytes == nil {
			log.Println("size 0", path)
			return nil
		}

		if containsSomeBytes(someBytes, bytesFilePathMap) {
			if err := os.Remove(path); err != nil {
				log.Fatal(err)
			}
		} else {
			dir, fileName := filepath.Split(path)
			subDir := filepath.Join(dir, strings.ReplaceAll(fileName, ".", "_"))
			toPath := filepath.Join(subDir, fileName)
			bytesFilePathMap[string(someBytes)] = []oldNewPath{{path, toPath}}
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	for _, oldNewPaths := range bytesFilePathMap {
		for _, oldNewPath := range oldNewPaths {
			log.Printf("[oldPath:%s] [newPath:%s]\n", oldNewPath.oldPath, oldNewPath.newPath)
		}
	}
	log.Printf("END  : %s\n", time.Now().Format(time.RFC3339))
}

func openDeDuplicationLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, deDuplicationLogFileName))
}
