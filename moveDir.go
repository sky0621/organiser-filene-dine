package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const moveDirLogFileName = "moveDir.log"

func moveDir(toDir string) {
	closeLogFile := openMoveDirLogFile(toDir)
	defer closeLogFile()

	log.Printf("START: %s\n", time.Now().Format(time.RFC3339))
	bytesFilePathMap := make(map[string][]oldNewPath)
	if err := filepath.WalkDir(toDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println("failed to WalkDir", err)
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			log.Println("failed to get directory info", err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		if !strings.Contains(path, dupDir) {
			return nil
		}

		dir, file := filepath.Split(path)
		log.Println(dir)
		log.Println(file)

		if file == ".DS_Store" {
			return nil
		}

		if strings.Contains(dir, ".organiser-filene-dine") {
			return nil
		}

		from := path
		to := filepath.Join(toDir, file)

		if err := renameFile(from, to); err != nil {
			log.Fatal(err)
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

func openMoveDirLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, moveDirLogFileName))
}
