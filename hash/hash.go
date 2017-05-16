package hash

import (
	"io"
	"path/filepath"
	"strings"

	"os"

	"time"

	"strconv"

	"github.com/cespare/xxhash"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/utils"
)

func SaveFileHash(file string, config itmconfig.ITMConfig) {
	absFile, _ := filepath.Abs(file)
	absSource, _ := filepath.Abs(config.SOURCE)
	relFile, _ := filepath.Rel(absSource, absFile)
	fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
	dirName := filepath.Join(config.DESTINATION, ".itm/files/"+fileName)
	now := strconv.FormatInt(time.Now().UTC().Unix(), 10)

	fileHash := GetFileHash(absFile)
	fileHashSet := GetFileHashSet(absFile, 2048)

	if config.USE_SSH {
		_, err := fsftp.Client.Stat(dirName)
		if err != nil {
			fsftp.Client.Mkdir(dirName)
		}
		fsftp.Client.Mkdir(filepath.Join(dirName, now))
	} else {
		_, err := os.Stat(dirName)
		if err != nil {
			os.Mkdir(dirName, os.ModePerm)
		}
		os.Mkdir(filepath.Join(dirName, now), os.ModePerm)
		utils.WriteLines([]string{fileHash}, filepath.Join(dirName, now, "hash.itmmi"))
		utils.WriteLines(fileHashSet, filepath.Join(dirName, now, "hashSet.itmi"))
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
