package main

import (
	"fmt"
	"os"
	"path/filepath"

	"time"

	"strconv"

	"io/ioutil"

	"github.com/slavikmanukyan/itm/fs"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/status"
	"github.com/slavikmanukyan/itm/utils"
)

func Restore(config itmconfig.ITMConfig, timestamp int64) {
	timestamps := hash.GetRemoteTimestamps(config)
	restorePoint, index := hash.GetClosestTime(timestamps, timestamp)

	day := time.Unix(restorePoint, 0)
	fmt.Println("Backing up to ", day.Format("02 Jan 2006 15:04"))

	added, deleted, changed, rest := status.GetStatus(config, restorePoint)
	if len(added) == 0 && len(deleted) == 0 && len(changed) == 0 {
		fmt.Println("Everything is up-to-date")
		return
	}

	for _, file := range added {
		os.Remove(filepath.Join(config.SOURCE, file))
	}

	if index == 0 {
		DoFullBackup(rest, config)
		fmt.Println("\nBackup completed!")
		return
	}

	timestamps = timestamps[:index+1]
	files := append(deleted, changed...)

	for _, file := range files {
		destination, firstTime := GetFileFirstAppearance(file, config, timestamps)
		fileSrc := destination
		relFile, err := filepath.Rel(config.DESTINATION, destination)
		if err == nil {
			fileSrc = relFile
		}
		fmt.Print("Backing up: ", fileSrc)
		os.Remove(filepath.Join(config.SOURCE, fileSrc))
		if config.USE_SSH {
			fsftp.CopyFileFrom(destination, filepath.Join(config.SOURCE, fileSrc))
		} else {
			fs.CopyFile(destination, filepath.Join(config.SOURCE, fileSrc))
		}
		if firstTime == index {
			continue
		}
		for _, times := range timestamps[firstTime+1:] {
			if utils.ExistsRemoteFile(filepath.Join(hash.GetFileMetaDestination(file, config, strconv.FormatInt(times, 10)), "slices"), config) {
				AddFileSlices(file, config, times)
			}
		}
	}

	fmt.Println("\nBackup completed!")
}

func AddFileSlices(file string, config itmconfig.ITMConfig, timestamp int64) {
	slicesDir := filepath.Join(hash.GetFileMetaDestination(file, config, strconv.FormatInt(timestamp, 10)), "slices")

	var entries []os.FileInfo
	if config.USE_SSH {
		entries, _ = fsftp.Client.ReadDir(slicesDir)
	} else {
		entries, _ = ioutil.ReadDir(slicesDir)
	}

	for _, slice := range entries {
		sliceData, count := utils.ReadRemoteFile(filepath.Join(slicesDir, slice.Name()), config)
		sliceIndex, _ := strconv.Atoi(slice.Name())
		fs.WriteFileSlice(filepath.Join(config.SOURCE, file), sliceIndex, 2048, sliceData, count)
	}
}

func GetFileFirstAppearance(file string, config itmconfig.ITMConfig, timestamps []int64) (string, int) {
	var err error

	for i := len(timestamps) - 1; i >= 0; i-- {
		metaPath := hash.GetFileMetaDestination(file, config, strconv.FormatInt(timestamps[i], 10))
		if config.USE_SSH {
			_, err = fsftp.Client.Stat(filepath.Join(metaPath, filepath.Base(file)))
		} else {
			_, err = os.Stat(filepath.Join(metaPath, file))
		}
		if err == nil {
			return filepath.Join(metaPath, file), i
		}
	}
	return filepath.Join(config.DESTINATION, file), 0
}

func DoFullBackup(ignoreFiles []string, config itmconfig.ITMConfig) {
	var newConfig itmconfig.ITMConfig
	newConfig.IGNORE = make(map[string]bool)
	newConfig.IGNORE[filepath.Join(config.DESTINATION, ".itm")] = true
	for _, file := range ignoreFiles {
		newConfig.IGNORE[file] = true
	}
	if config.USE_SSH {
		fsftp.CopyDirFrom(config.DESTINATION, config.SOURCE, newConfig, func(file string) {
			fileSrc := file
			relFile, err := filepath.Rel(config.DESTINATION, file)
			if err == nil {
				fileSrc = relFile
			}
			if config.IS_TERMINAL {
				utils.ClearLine()
			}
			fmt.Print("Backing up: ", fileSrc)
		})
	} else {
		fs.CopyDir(config.DESTINATION, config.SOURCE, newConfig, "")
	}
}
