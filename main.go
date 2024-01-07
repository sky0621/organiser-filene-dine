package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

const fileListName = "fileList.txt"

func main() {
	if len(os.Args) < 3 {
		os.Exit(-1)
	}

	fromDir := os.Args[1]
	toDir := os.Args[2]

	fileList, err := os.Create(filepath.Join(toDir, fileListName))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fileList.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	existsSet := mapset.NewSet[string]()

	fileListScanner := bufio.NewScanner(fileList)
	for fileListScanner.Scan() {
		existsSet.Add(fileListScanner.Text())
	}

	if err := filepath.WalkDir(fromDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(path)

		fi, err := d.Info()
		if err != nil {
			log.Println(err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		if strings.HasPrefix(fi.Name(), ".") {
			return nil
		}

		return exec(path, existsSet, toDir, fi)
	}); err != nil {
		log.Fatal(err)
	}

	for _, s := range existsSet.ToSlice() {
		_, err := fileList.WriteString(fmt.Sprintf("%s\n", s))
		if err != nil {
			log.Println(err)
		}
	}
}

func exec(path string, existsSet mapset.Set[string], toDir string, fi fs.FileInfo) error {
	if !addExists(existsSet, fi) {
		return nil
	}

	/*
	 * Output Directory
	 */
	outDirName := getOutputDirName(fi.ModTime())
	if err := createOutputDir(toDir, outDirName); err != nil {
		log.Println(err)
		return nil
	}

	/*
	 * Output File
	 */
	outFile, err := os.Create(filepath.Join(toDir, outDirName, createOutFileName(fi)))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	/*
	 * Input File
	 */
	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	/*
	 * Copy
	 */
	if _, err := io.Copy(outFile, f); err != nil {
		log.Println(err)
		return nil
	}

	return nil
}

func addExists(existsSet mapset.Set[string], fi fs.FileInfo) bool {
	key := fmt.Sprintf("%s%s%d", fi.Name(), formatModTime(fi.ModTime()), fi.Size())
	if existsSet.Contains(key) {
		log.Println("=====================")
		log.Println("EXISTS:", key)
		log.Println("=====================")
		return false
	}
	existsSet.Add(key)
	return true
}

func formatModTime(modTime time.Time) string {
	return modTime.Format("2006-01-02T15h04m05s")
}

func createOutFileName(fi fs.FileInfo) string {
	return fmt.Sprintf("%s_%d%s", formatModTime(fi.ModTime()), fi.Size(), filepath.Ext(fi.Name()))
}

func getOutputDirName(modTime time.Time) string {
	return modTime.Format("200601")
}

func createOutputDir(toDir string, outDirName string) error {
	outDir := filepath.Join(toDir, outDirName)
	if f, err := os.Stat(outDir); os.IsNotExist(err) || !f.IsDir() {
		return os.Mkdir(outDir, fs.ModePerm)
	}
	return nil
}
