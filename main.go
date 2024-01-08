package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	operationAll          = 9
	operationPrepare      = 1
	operationCreateOutDir = 2
	operationCopy         = 3
)

const copyListFileName = "copyList.txt"
const seps = "#-#-#$%&**&%$#-#-#"

const outputDirSetFileName = "outputDirSet.txt"

const errorListName = "errorList.txt"

type Config struct {
	FromDir             string   `yaml:"fromDir"`
	ToDir               string   `yaml:"toDir"`
	TargetExts          string   `yaml:"targetExts"`
	TargetDocumentsExts []string `yaml:"targetDocumentsExts"`
	TargetPicturesExts  []string `yaml:"targetPicturesExts"`
	TargetMusicsExts    []string `yaml:"targetMusicsExts"`
	TargetMoviesExts    []string `yaml:"targetMoviesExts"`
	Rename              bool     `yaml:"rename"`
	Operation           int      `yaml:"operation"`
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

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config/")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}

	targetExts := getTargetExts(cfg)

	/****************************************************************
	 * create copy list
	 */
	outputDirSet := mapset.NewSet[string]()
	if cfg.Operation == operationPrepare || cfg.Operation == operationAll {
		prepareLogFile, err := os.OpenFile("organiser-filene-dine-prepare.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := prepareLogFile.Close(); err != nil {
				log.Println(err)
			}
		}()

		log.SetOutput(prepareLogFile)

		copyListFile, err := os.Create(filepath.Join(cfg.ToDir, copyListFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := copyListFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		if err := filepath.WalkDir(cfg.FromDir, func(path string, d os.DirEntry, err error) error {
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

			return prepare(path, cfg.ToDir, fi, cfg.Rename, copyListFile, outputDirSet)
		}); err != nil {
			log.Fatal(err)
		}

		outputDirSetFile, err := os.Create(filepath.Join(cfg.ToDir, outputDirSetFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := outputDirSetFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		for _, outputDir := range outputDirSet.ToSlice() {
			_, err := outputDirSetFile.WriteString(fmt.Sprintf("%s\n", filepath.Join(cfg.ToDir, outputDir)))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	/****************************************************************
	 * create output directory
	 */
	if cfg.Operation == operationCreateOutDir || cfg.Operation == operationAll {
		outputdirLogFile, err := os.OpenFile("organiser-filene-dine-outputdir.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := outputdirLogFile.Close(); err != nil {
				log.Println(err)
			}
		}()

		log.SetOutput(outputdirLogFile)

		outputDirSetFile, err := os.Open(filepath.Join(cfg.ToDir, outputDirSetFileName))
		if err != nil {
			log.Printf("failed to open %s\n", outputDirSetFileName)
			log.Fatal(err)
		}
		defer func() {
			if err := outputDirSetFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		outputDirSetFileScanner := bufio.NewScanner(outputDirSetFile)
		for outputDirSetFileScanner.Scan() {
			dirPath := outputDirSetFileScanner.Text()
			if err := os.Mkdir(dirPath, fs.ModePerm); err != nil {
				log.Printf("failed to mkdir: %s", dirPath)
				log.Fatal(err)
			}
			log.Printf("created: %s\n", dirPath)
		}
	}

	/****************************************************************
	 * copy
	 */
	if cfg.Operation == operationCopy || cfg.Operation == operationAll {
		copyLogFile, err := os.OpenFile("organiser-filene-dine-copy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			if err := copyLogFile.Close(); err != nil {
				log.Println(err)
			}
		}()

		log.SetOutput(copyLogFile)
		log.Printf("START: %s\n", time.Now().Format(time.RFC3339))

		copyListFile, err := os.Open(filepath.Join(cfg.ToDir, copyListFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := copyListFile.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		errorList, err := os.Create(filepath.Join(cfg.ToDir, errorListName))
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := errorList.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		cpuNum := runtime.NumCPU()
		log.Printf("NumCPU: %d\n", cpuNum)

		// ★ 同時実行 goroutine 数の制御のためにチャネル用意
		semaphore := make(chan struct{}, cpuNum*4)

		wg := &sync.WaitGroup{}

		copyListFileScanner := bufio.NewScanner(copyListFile)
		for copyListFileScanner.Scan() {
			semaphore <- struct{}{}
			wg.Add(1)

			line := copyListFileScanner.Text()
			fromTo := strings.Split(line, seps)
			go func() {
				err := copyFile(fromTo[0], fromTo[1], errorList, semaphore, wg)
				if err != nil {
					log.Println(err)
				}
			}()
		}

		wg.Wait()
		log.Printf("END  : %s\n", time.Now().Format(time.RFC3339))
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
		log.Println("[[[ failed to copyList.WriteString ]]]")
		log.Println(err)
		return err
	}

	return nil
}

func copyFile(fromPath string, toPath string, errorList *os.File, semaphore chan struct{}, wg *sync.WaitGroup) error {
	defer func() {
		<-semaphore // 処理後にチャネルから値を抜き出さないと、次の goroutine が起動できない
	}()
	defer wg.Done()

	fromFile, err := os.Open(fromPath)
	if err != nil {
		log.Println("[[[ failed to open fromFile ]]]")
		log.Println(err)
		_, err = errorList.WriteString(fromPath + "\n")
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	defer func() {
		if err := fromFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	toFile, err := os.Create(toPath)
	if err != nil {
		log.Println("[[[ failed to open toFile ]]]")
		log.Println(err)
		_, err = errorList.WriteString(toPath + "\n")
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	defer func() {
		if err := toFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	if _, err := io.Copy(toFile, fromFile); err != nil {
		log.Println("[[[ failed to copy ]]]")
		log.Println(err)
		_, err = errorList.WriteString(fmt.Sprintf("[from:%s] [to:%s]\n", fromPath, toPath))
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	log.Printf("copied:[from:%s] [to:%s]\n", fromPath, toPath)

	return nil
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

func contains(strs []string, s string) bool {
	ls := strings.ToLower(s)
	for _, str := range strs {
		if str == ls {
			return true
		}
	}
	return false
}
