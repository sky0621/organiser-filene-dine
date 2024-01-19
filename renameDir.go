package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const renameDirLogFileName = "renameDir.log"

func renameDir(toDir string) {
	closeLogFile := openRenameDirLogFile(toDir)
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

		if !fi.IsDir() {
			return nil
		}

		dir, file := filepath.Split(path)
		log.Println(dir)
		log.Println(file)

		if strings.Contains(file, "___Volumes___HD-LCU3___") {
			newSubDir := strings.Replace(file, "___Volumes___HD-LCU3___", "", -1)
			if err := renameFile(path, filepath.Join(toDir, newSubDir)); err != nil {
				log.Fatal(err)
			}
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

func openRenameDirLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, renameDirLogFileName))
}
