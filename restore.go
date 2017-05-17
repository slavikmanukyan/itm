package main

import (
	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
)

func Restore(config itmconfig.ITMConfig, timestamp int64) {
	timestamps := hash.GetRemoteTimestamps(config)
	restorePoint := hash.GetClosestTime(timestamps, timestamp)

	timestamps = timestamps[1:]

	DoFullBackup(config)
	for i := 0; timestamps[i] <= restorePoint; i++ {

	}
}

func DoFullBackup(config itmconfig.ITMConfig) {

}
