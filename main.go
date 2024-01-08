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
	if len(os.Args) < 5 {
		os.Exit(-1)
	}

	fromDir := os.Args[1]
	toDir := os.Args[2]
	rename := os.Args[3]
	isRename := true
	if rename == "0" {
		isRename = false
	}
	log.Println(isRename)
	targetExts := strings.Split(os.Args[4], ",")

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
			log.Println("[ 01 ]")
			log.Println(err)
			return err
		}

		fi, err := d.Info()
		if err != nil {
			log.Println("[ 02 ]")
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

		return exec(path, existsSet, toDir, fi, isRename, fileList)
	}); err != nil {
		log.Fatal(err)
	}

}

func exec(path string, existsSet mapset.Set[string], toDir string, fi fs.FileInfo, isRename bool, fileList *os.File) error {
	name := fi.Name()
	createdTime := getCreatedTime(fi)
	size := fi.Size()

	element := createExistsSetElement(name, formatTime(createdTime), size)
	if !addExists(existsSet, element) {
		return nil
	}

	/*
	 * Output Directory
	 */
	outDirName := getOutputDirName(path)
	if err := createOutputDir(toDir, outDirName); err != nil {
		log.Println("[[[ 01 ]]]")
		log.Println(err)
		return nil
	}

	/*
	 * Output File
	 */
	outFileName := createOutFileName(createdTime, size, getExt(name))
	if !isRename {
		outFileName = name
	}
	outFile, err := os.Create(filepath.Join(toDir, outDirName, outFileName))
	if err != nil {
		log.Println("[[[ 02 ]]]")
		log.Println(err)
		return nil
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Println("[[[ 03 ]]]")
			log.Println(err)
		}
	}()

	/*
	 * Input File
	 */
	f, err := os.Open(path)
	if err != nil {
		log.Println("[[[ 04 ]]]")
		log.Println(err)
		return nil
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("[[[ 05 ]]]")
			log.Println(err)
		}
	}()

	/*
	 * Copy
	 */
	if _, err := io.Copy(outFile, f); err != nil {
		log.Println("[[[ 06 ]]]")
		log.Println(err)
		return nil
	}

	_, err = fileList.WriteString(fmt.Sprintf("%s\n", element))
	if err != nil {
		log.Println(err)
	}

	return nil
}

func createExistsSetElement(name string, birthTime string, size int64) string {
	return fmt.Sprintf("%s%s%d", name, birthTime, size)
}

func addExists(existsSet mapset.Set[string], element string) bool {
	if existsSet.Contains(element) {
		log.Println("[EXISTS]:", element)
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

func getOutputDirName(path string) string {
	dir, _ := filepath.Split(path)
	//log.Println(dir)
	dirs := strings.Split(dir, "/")
	//dirs := filepath.SplitList(dir)
	//log.Println(dirs)
	//log.Println(len(dirs))
	if len(dirs) < 2 {
		return "root"
	}
	ret := dirs[len(dirs)-2]
	//log.Println(ret)
	return ret
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
