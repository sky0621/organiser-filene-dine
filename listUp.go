package main

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
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

	createOutputExtsDirectory(cfg)

	allTargetExts := getAllTargetExts(cfg)

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

		if cfg.TargetExts == TargetExtsAll {
			log.Println(path)
			return prepare(path, fi, cfg, copyListFile, outputDirSet)
		}

		if cfg.TargetExts == TargetExtsOthers {
			if contains(allTargetExts, getExt(fi.Name())) {
				log.Println("[NOT_TARGET]", fi.Name())
				return nil
			}
		} else {
			if !contains(getTargetExts(cfg), getExt(fi.Name())) {
				log.Println("[NOT_TARGET]", fi.Name())
				return nil
			}
		}

		log.Println(path)
		return prepare(path, fi, cfg, copyListFile, outputDirSet)
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

func prepare(fromPath string, fi fs.FileInfo, cfg Config, copyList *os.File, outputDirSet mapset.Set[string]) error {
	outDirName := getOutputDirName(fromPath)
	extsDir := getOutputExtsDirectoryName(getExt(fi.Name()), cfg)
	outputDirSet.Add(filepath.Join(extsDir, outDirName))

	outFileName := ""
	if cfg.Rename {
		outFileName = createOutFileName(getCreatedTime(fi), uuid.NewString(), fi.Name())
	} else {
		outFileName = fi.Name()
	}

	toPath := filepath.Join(cfg.ToDir, extsDir, outDirName, outFileName)

	_, err := copyList.WriteString(fmt.Sprintf("%s%s%s\n", fromPath, seps, toPath))
	if err != nil {
		log.Println("[[[ failed to copyList.WriteString ]]]", err)
		return err
	}

	return nil
}

func createOutputExtsDirectory(cfg Config) {
	if cfg.TargetExts == TargetExtsAll {
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsDocuments))
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsImages))
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsMusics))
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsVideos))
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsOthers))
		return
	}

	switch cfg.TargetExts {
	case TargetExtsDocuments:
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsDocuments))
	case TargetExtsImages:
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsImages))
	case TargetExtsMusics:
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsMusics))
	case TargetExtsVideos:
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsVideos))
	default:
		createDirectory(filepath.Join(cfg.ToDir, TargetExtsOthers))
	}
}

func getOutputExtsDirectoryName(ext string, cfg Config) string {
	if contains(cfg.TargetDocumentsExts, ext) {
		return TargetExtsDocuments
	}
	if contains(cfg.TargetImagesExts, ext) {
		return TargetExtsImages
	}
	if contains(cfg.TargetMusicsExts, ext) {
		return TargetExtsMusics
	}
	if contains(cfg.TargetVideosExts, ext) {
		return TargetExtsVideos
	}
	return TargetExtsOthers
}

func getTargetExts(cfg Config) []string {
	switch cfg.TargetExts {
	case TargetExtsDocuments:
		return cfg.TargetDocumentsExts
	case TargetExtsImages:
		return cfg.TargetImagesExts
	case TargetExtsMusics:
		return cfg.TargetMusicsExts
	case TargetExtsVideos:
		return cfg.TargetVideosExts
	}
	return nil
}

func getAllTargetExts(cfg Config) []string {
	s1 := append(cfg.TargetDocumentsExts, cfg.TargetImagesExts...)
	s2 := append(s1, cfg.TargetMusicsExts...)
	s3 := append(s2, cfg.TargetVideosExts...)
	return s3
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
	ret := strings.ReplaceAll(dir, "/", "___")
	if ret == "" {
		ret = "root"
	}
	//dirs := strings.Split(dir, "/")
	//if len(dirs) < 2 {
	//	return "root"
	//}
	//ret := dirs[len(dirs)-2]
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

func createOutFileName(createdTime time.Time, uuid string, fileName string) string {
	return fmt.Sprintf("%s_%s_%s", formatTime(createdTime), uuid, fileName)
}
