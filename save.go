package main

import (
	"strings"
	"time"

	"strconv"

	"path/filepath"

	"fmt"

	"github.com/slavikmanukyan/itm/fs"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/status"
	"github.com/slavikmanukyan/itm/utils"
)

func Save(config itmconfig.ITMConfig) {
	added, deleted, changed, rest := status.GetStatus(config, 0)
	if len(added) == 0 && len(changed) == 0 && len(deleted) == 0 {
		fmt.Println("Everything is up-to-date")
		return
	}

	timestamps := hash.GetRemoteTimestamps(config)
	last, _ := hash.GetClosestTime(timestamps, time.Now().UTC().Unix())

	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	dirName := filepath.Join(config.DESTINATION, ".itm", "files", timestamp)
	utils.CreateRemoteDir(dirName, config)

	for _, file := range append(append(added, changed...), rest...) {
		hash.SaveFileHash(filepath.Join(config.SOURCE, file), config, timestamp)
	}

	fmt.Println("Calculating difference...")
	for _, file := range changed {
		changedSlices := make(map[string][]byte)

		hashSet1 := hash.GetFileHashSet(filepath.Join(config.SOURCE, file), 2048)
		hashSet2 := hash.GetRemoteHashSet(file, config, strconv.FormatInt(last, 10))
		for i, hash := range hashSet1 {
			if hash != hashSet2[i] {
				slice := fs.ReadFileSlice(filepath.Join(config.SOURCE, file), i, 2048)
				changedSlices[hash] = slice
			}
		}

		absFile, _ := filepath.Abs(file)
		absSource, _ := filepath.Abs(config.SOURCE)
		relFile, _ := filepath.Rel(absSource, absFile)
		fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
		dirName := filepath.Join(config.DESTINATION, ".itm/files/", timestamp, fileName, "slices")

		utils.CreateRemoteDir(dirName, config)
		for hash, data := range changedSlices {
			utils.WriteRemoteFile(filepath.Join(dirName, hash), config, data)
		}
	}

	for _, file := range added {
		absFile, _ := filepath.Abs(file)
		absSource, _ := filepath.Abs(config.SOURCE)
		relFile, _ := filepath.Rel(absSource, absFile)
		fileName := strings.Replace(filepath.ToSlash(relFile), "/", "-", -1)
		filePath := filepath.Join(config.DESTINATION, ".itm", "files", timestamp, fileName, filepath.Base(file))
		if config.USE_SSH {
			fsftp.CopyFileTo(filepath.Join(config.SOURCE, file), filePath)
		} else {
			fs.CopyFile(filepath.Join(config.SOURCE, file), filePath)
		}
	}

	fmt.Println("Saved: ")
	if len(added) > 0 {
		fmt.Println("      ", len(added), "new files")
	}
	if len(changed) > 0 {
		fmt.Println("       updated", len(changed), "files")
	}
}
