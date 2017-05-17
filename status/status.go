package status

import (
	"io/ioutil"
	"path/filepath"

	"time"

	"strconv"

	"github.com/slavikmanukyan/itm/fs"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/utils"
)

func GetStatus(config itmconfig.ITMConfig) (added []string, deleted []string, changed []string) {
	currentHashes := GetCurrentHashes(config)
	timestamps := hash.GetRemoteTimestamps(config)
	now := hash.GetClosestTime(timestamps, time.Now().UTC().Unix())

	remoteHashes := GetRemoteHashes(config, strconv.FormatInt(now, 10))

	for key, val := range currentHashes {
		if remoteHashes[key] == "" {
			added = append(added, key)
		} else if remoteHashes[key] != val {
			changed = append(changed, key)
		}
	}
	for key := range remoteHashes {
		if currentHashes[key] == "" {
			deleted = append(deleted, key)
		}
	}
	return
}

func GetCurrentHashes(config itmconfig.ITMConfig) map[string]string {
	hashes := make(map[string]string)
	files, _ := fs.WalkAll(config.SOURCE, config)
	for _, file := range files {
		relFile, _ := filepath.Rel(config.SOURCE, file)
		hashes[relFile] = hash.GetFileHash(file)
	}
	return hashes
}

func GetRemoteHashes(config itmconfig.ITMConfig, timestamp string) map[string]string {
	hashes := make(map[string]string)
	destination := filepath.Join(config.DESTINATION, ".itm", "files", timestamp)
	if config.USE_SSH {
		files, _ := fsftp.Client.ReadDir(destination)
		for _, file := range files {
			hf, _ := utils.ReadLinesRemote(filepath.Join(destination, file.Name(), "hash.itmi"), fsftp.Client)
			hashes[hf[1]] = hf[0]
		}
	} else {
		files, _ := ioutil.ReadDir(destination)
		for _, file := range files {
			hf, _ := utils.ReadLines(filepath.Join(destination, file.Name(), "hash.itmi"))
			hashes[hf[1]] = hf[0]
		}
	}
	return hashes
}
