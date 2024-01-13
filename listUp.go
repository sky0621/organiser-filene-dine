package main

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const copyListFileName = "copyList.txt"
const listUpLogFileName = "listUp.log"

func listUp(cfg Config) {
	outputDirSet := mapset.NewSet[string]()

	closeLogFile := openListUpLogFile(cfg.ToDir)
	defer closeLogFile()

	copyListFile, closeCopyListFile := openCopyListFile(cfg.ToDir)
	defer closeCopyListFile()

	if err := filepath.WalkDir(cfg.FromDir, func(path string, d os.DirEntry, err error) error {
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

		if !contains(getTargetExts(cfg), getExt(fi.Name())) {
			log.Println("[NOT_TARGET]", fi.Name())
			return nil
		}

		log.Println(path)

		return prepare(path, cfg.ToDir, fi, cfg.Rename, copyListFile, outputDirSet)
	}); err != nil {
		log.Fatal(err)
	}

	outputDirSetFile, closeOutputDirSetFile := openFile(getOutputDirSetFilePath(cfg.ToDir))
	defer closeOutputDirSetFile()

	for _, outputDir := range outputDirSet.ToSlice() {
		log.Println(outputDir)
		_, err := outputDirSetFile.WriteString(fmt.Sprintf("%s\n", filepath.Join(cfg.ToDir, outputDir)))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func prepare(fromPath string, toDir string, fi fs.FileInfo, rename bool, copyList *os.File, outputDirSet mapset.Set[string]) error {
	outDirName := getOutputDirName(fromPath)
	outputDirSet.Add(outDirName)

	outFileName := ""
	if rename {
		outFileName = createOutFileName(getCreatedTime(fi), fi.Size(), getExt(fi.Name()))
	} else {
		outFileName = fi.Name()
	}

	toPath := filepath.Join(toDir, outDirName, outFileName)

	_, err := copyList.WriteString(fmt.Sprintf("%s%s%s\n", fromPath, seps, toPath))
	if err != nil {
		log.Println("[[[ failed to copyList.WriteString ]]]", err)
		return err
	}

	return nil
}

func getTargetExts(conf Config) []string {
	switch conf.TargetExts {
	case "Documents":
		return conf.TargetDocumentsExts
	case "Pictures":
		return conf.TargetPicturesExts
	case "Musics":
		return conf.TargetMusicsExts
	case "Movies":
		return conf.TargetMoviesExts
	}
	return nil
}

func openListUpLogFile(rootPath string) CloseFunc {
	return setupLog(filepath.Join(rootPath, metaDir, listUpLogFileName))
}

func openCopyListFile(rootPath string) (*os.File, CloseFunc) {
	copyListFilePath := getCopyListFilePath(rootPath)
	copyListFileBackupPath := getCopyListBackupFilePath(rootPath)
	if err := renameFile(copyListFilePath, copyListFileBackupPath); err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Fatal(err)
		}
	}
	return openFile(copyListFilePath)
}

func getOutputDirName(path string) string {
	dir, _ := filepath.Split(path)
	dirs := strings.Split(dir, "/")
	if len(dirs) < 2 {
		return "root"
	}
	ret := dirs[len(dirs)-2]
	return ret
}

func toTime(ts syscall.Timespec) time.Time {
	return time.Unix(ts.Sec, ts.Nsec)
}

func getCreatedTime(fi fs.FileInfo) time.Time {
	statT := fi.Sys().(*syscall.Stat_t)
	return toTime(statT.Birthtimespec)
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02T15h04m05s")
}

func createOutFileName(createdTime time.Time, size int64, ext string) string {
	return fmt.Sprintf("%s_%d%s", formatTime(createdTime), size, ext)
}
