package main

import (
	"bufio"
	"io/fs"
	"log"
	"os"
)

func createOutputDir(cfg Config) {
	closeLogFile := setupLog("createOutputDir")
	defer closeLogFile()

	outputDirSetFile, closeOutputDirSetFile := openOutputDirSetFile(cfg.ToDir)
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
