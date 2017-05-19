package hash

import (
	"io"
	"path/filepath"
	"strings"

	"os"

	"strconv"

	"io/ioutil"

	"math"

	"sort"

	"github.com/cespare/xxhash"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/utils"
)

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

func SaveFileHash(file string, config itmconfig.ITMConfig, timestamp string) {
	absFile, _ := filepath.Abs(file)
	absSource, _ := filepath.Abs(config.SOURCE)
	relFile, _ := filepath.Rel(absSource, absFile)
	fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
	dirName := filepath.Join(config.DESTINATION, ".itm/files/", timestamp, fileName)

	fileHash := GetFileHash(absFile)
	fileHashSet := GetFileHashSet(absFile, 2048)

	if config.USE_SSH {
		_, err := fsftp.Client.Stat(dirName)
		if err != nil {
			fsftp.Client.Mkdir(dirName)
		}
		utils.WriteLinesRemote([]string{fileHash, relFile}, filepath.Join(dirName, "hash.itmi"), fsftp.Client)
		utils.WriteLinesRemote(fileHashSet, filepath.Join(dirName, "hashSet.itmi"), fsftp.Client)
	} else {
		_, err := os.Stat(dirName)
		if err != nil {
			os.Mkdir(dirName, os.ModePerm)
		}
		utils.WriteLines([]string{fileHash, relFile}, filepath.Join(dirName, "hash.itmi"))
		utils.WriteLines(fileHashSet, filepath.Join(dirName, "hashSet.itmi"))
	}
}

func GetFileHashSet(file string, size int) []string {
	fi, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	// make a buffer to keep chunks that are read
	buf := make([]byte, size)
	var hashSet []string
	for {
		// read a chunk
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		h := xxhash.New()
		h.Write(buf[:n])
		hashSet = append(hashSet, strconv.FormatUint(h.Sum64(), 16))
	}
	return hashSet
}

func GetFileHash(file string) string {
	fi, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	h := xxhash.New()
	if _, err := io.Copy(h, fi); err != nil {
		panic(err)
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

func GetRemoteTimestamps(config itmconfig.ITMConfig) []int64 {
	var timeFolders []os.FileInfo
	var err error
	destination := filepath.Join(config.DESTINATION, ".itm", "files")
	if config.USE_SSH {
		timeFolders, err = fsftp.Client.ReadDir(destination)
	} else {
		timeFolders, err = ioutil.ReadDir(destination)
	}
	if err != nil {
		panic(err)
	}
	var timestamps int64arr
	for _, dir := range timeFolders {
		bTime, err := strconv.ParseInt(dir.Name(), 10, 64)
		if err == nil {
			timestamps = append(timestamps, bTime)
		}
	}
	sort.Sort(timestamps)
	return timestamps
}

func GetClosestTime(timestamps int64arr, byTime int64) (int64, int) {
	sort.Sort(timestamps)
	if byTime <= timestamps[0] {
		return timestamps[0], 0
	}
	if byTime >= timestamps[len(timestamps)-1] {
		return timestamps[len(timestamps)-1], len(timestamps) - 1
	}
	for i := 0; i < len(timestamps)-1; i++ {
		if math.Abs(float64(timestamps[i]-byTime)) <= math.Abs(float64(timestamps[i+1]-byTime)) {
			return timestamps[i], i
		}
	}
	return timestamps[len(timestamps)-1], len(timestamps) - 1
}

func GetFileMetaDestination(file string, config itmconfig.ITMConfig, timestamp string) string {
	absFile, _ := filepath.Abs(file)
	absSource, _ := filepath.Abs(config.SOURCE)
	relFile, _ := filepath.Rel(absSource, absFile)
	fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
	src := filepath.Join(config.DESTINATION, ".itm", "files", timestamp, fileName)

	return src
}

func GetRemoteHash(file string, config itmconfig.ITMConfig, timestamp string) (string, string) {
	absFile, _ := filepath.Abs(file)
	absSource, _ := filepath.Abs(config.SOURCE)
	relFile, _ := filepath.Rel(absSource, absFile)
	fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
	src := filepath.Join(config.DESTINATION, ".itm", "files", timestamp, fileName, "hash.itmi")

	var hash []string
	var err error

	if config.USE_SSH {
		hash, err = utils.ReadLinesRemote(src, fsftp.Client)
	} else {
		hash, err = utils.ReadLines(src)
	}
	if err != nil {
		panic(err)
	}
	return hash[0], hash[1]
}

func GetRemoteHashSet(file string, config itmconfig.ITMConfig, timestamp string) []string {
	absFile, _ := filepath.Abs(file)
	absSource, _ := filepath.Abs(config.SOURCE)
	relFile, _ := filepath.Rel(absSource, absFile)
	fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
	src := filepath.Join(config.DESTINATION, ".itm", "files", timestamp, fileName, "hashSet.itmi")

	var hashSet []string
	var err error

	if config.USE_SSH {
		hashSet, err = utils.ReadLinesRemote(src, fsftp.Client)
	} else {
		hashSet, err = utils.ReadLines(src)
	}
	if err != nil {
		panic(err)
	}
	return hashSet
}
