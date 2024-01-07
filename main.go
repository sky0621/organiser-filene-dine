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
	"syscall"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

const fileListName = "fileList.txt"

func main() {
	if len(os.Args) < 4 {
		os.Exit(-1)
	}

	fromDir := os.Args[1]
	toDir := os.Args[2]
	targetExts := strings.Split(os.Args[3], ",")

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

		fi, err := d.Info()
		if err != nil {
			log.Println(err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		if !contains(targetExts, getExt(fi.Name())) {
			log.Println("[NOT_TARGET]", fi.Name())
			return nil
		}

		log.Println(path)

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
	name := fi.Name()
	createdTime := getCreatedTime(fi)
	size := fi.Size()

	if !addExists(existsSet, createExistsSetElement(name, formatTime(createdTime), size)) {
		return nil
	}

	/*
	 * Output Directory
	 */
	outDirName := getOutputDirName(createdTime)
	if err := createOutputDir(toDir, outDirName); err != nil {
		log.Println(err)
		return nil
	}

	/*
	 * Output File
	 */
	outFileName := createOutFileName(createdTime, size, getExt(name))
	outFile, err := os.Create(filepath.Join(toDir, outDirName, outFileName))
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

func createExistsSetElement(name string, birthTime string, size int64) string {
	return fmt.Sprintf("%s%s%d", name, birthTime, size)
}

func addExists(existsSet mapset.Set[string], element string) bool {
	if existsSet.Contains(element) {
		log.Println("=====================")
		log.Println("EXISTS:", element)
		log.Println("=====================")
		return false
	}
	existsSet.Add(element)
	return true
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15h04m05s")
}

func getExt(fileName string) string {
	return filepath.Ext(fileName)
}

func createOutFileName(createdTime time.Time, size int64, ext string) string {
	return fmt.Sprintf("%s_%d%s", formatTime(createdTime), size, ext)
}

func getOutputDirName(t time.Time) string {
	return t.Format("200601")
}

func createOutputDir(toDir string, outDirName string) error {
	outDir := filepath.Join(toDir, outDirName)
	if f, err := os.Stat(outDir); os.IsNotExist(err) || !f.IsDir() {
		return os.Mkdir(outDir, fs.ModePerm)
	}
	return nil
}

func toTime(ts syscall.Timespec) time.Time {
	return time.Unix(ts.Sec, ts.Nsec)
}

func getCreatedTime(fi fs.FileInfo) time.Time {
	statT := fi.Sys().(*syscall.Stat_t)
	return toTime(statT.Birthtimespec)
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
