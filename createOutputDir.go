package main

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const createOutputDirLogFileName = "createOutputDir.log"

func createOutputDir(cfg Config) {
	closeLogFile := openCreateOutputDirLogFile(cfg.ToDir)
	defer closeLogFile()

	outputDirSetFile, closeOutputDirSetFile := open(getOutputDirSetFilePath(cfg.ToDir))
	defer closeOutputDirSetFile()

	outputDirSetFileScanner := bufio.NewScanner(outputDirSetFile)
	for outputDirSetFileScanner.Scan() {
		dirPath := outputDirSetFileScanner.Text()
		if err := os.Mkdir(dirPath, fs.ModePerm); err != nil {
			log.Printf("failed to mkdir %s: %s\n", dirPath, err.Error())
			continue
		}
		log.Printf("created: %s\n", dirPath)
	}
}

func openCreateOutputDirLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, createOutputDirLogFileName))
}
