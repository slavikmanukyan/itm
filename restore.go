package main

import (
	"fmt"
	"os"
	"path/filepath"

	"time"

	"github.com/slavikmanukyan/itm/fs"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/status"
)

func Restore(config itmconfig.ITMConfig, timestamp int64) {
	timestamps := hash.GetRemoteTimestamps(config)
	restorePoint, index := hash.GetClosestTime(timestamps, timestamp)

	day := time.Unix(restorePoint, 0)
	fmt.Println("Backing up to ", day.Format("02 Jan 2006 15:04"))

	added, deleted, changed, _ := status.GetStatus(config, restorePoint)
	if len(added) == 0 && len(deleted) == 0 && len(changed) == 0 {
		fmt.Println("Everything is up-to-date")
		return
	}

	for _, file := range added {
		os.Remove(filepath.Join(config.SOURCE, file))
	}

	if index == 0 {
		DoFullBackup(config, timestamps[0])
		return
	}

	timestamps = timestamps[1:index]
	for i := 0; timestamps[i] <= restorePoint; i++ {

	}
}

func GetFileFirstAppearance(file string, config itmconfig.ITMConfig, timestamps []int64) {

}

func DoFullBackup(config itmconfig.ITMConfig, first int64) {
	var newConfig itmconfig.ITMConfig
	newConfig.IGNORE = make(map[string]bool)
	newConfig.IGNORE[filepath.Join(config.DESTINATION, ".itm")] = false
	if config.USE_SSH {
		fsftp.CopyDirFrom(config.DESTINATION, config.SOURCE, newConfig)
	} else {
		fs.CopyDir(config.DESTINATION, config.SOURCE, newConfig, "")
	}
}
