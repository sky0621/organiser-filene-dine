package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const checkDuplicationLogFileName = "checkDuplication.log"
const dupDir = "__duplicated__"

func checkDuplication(toDir string) {
	closeLogFile := openCheckDuplicationLogFile(toDir)
	defer closeLogFile()

	createDirectory(filepath.Join(toDir, dupDir))

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

		if strings.Contains(path, dupDir) {
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
			alreadyPaths := bytesFilePathMap[string(someBytes)]
			onPath := alreadyPaths[0]
			alreadyDir, _ := filepath.Split(onPath.newPath)

			if len(alreadyPaths) == 1 {
				createDirectory(alreadyDir)
				log.Printf("onPath.newPath: %s\n", onPath.newPath)
				if err := renameFile(onPath.oldPath, onPath.newPath); err != nil {
					log.Fatal(err)
				}
			}

			toFileName := createWithSubDirFileName(path)
			log.Printf("(in contains) toFileName: %s\n", toFileName)
			toPath := filepath.Join(alreadyDir, toFileName)
			if err := renameFile(path, toPath); err != nil {
				log.Fatal(err)
			}

			bytesFilePathMap[string(someBytes)] = append(bytesFilePathMap[string(someBytes)], oldNewPath{path, toPath})

			return nil
		}

		toFileName := createWithSubDirFileName(path)
		log.Printf("toFileName: %s\n", toFileName)
		toPath := filepath.Join(toDir, dupDir, uuid.NewString(), toFileName)
		bytesFilePathMap[string(someBytes)] = []oldNewPath{{path, toPath}}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	//for _, oldNewPaths := range bytesFilePathMap {
	//	for _, oldNewPath := range oldNewPaths {
	//		log.Printf("[oldPath:%s] [newPath:%s]\n", oldNewPath.oldPath, oldNewPath.newPath)
	//	}
	//}
	log.Printf("END  : %s\n", time.Now().Format(time.RFC3339))
}

func openCheckDuplicationLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, checkDuplicationLogFileName))
}

func createWithSubDirFileName(path string) string {
	dir, fileName := filepath.Split(path)
	dirs := strings.Split(dir, "/")
	subDir := dirs[len(dirs)-1]
	if subDir == "" {
		subDir = dirs[len(dirs)-2]
	}
	return fmt.Sprintf("%s____%s", subDir, fileName)
}
