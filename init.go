package main

import (
	"errors"
	"path/filepath"

	"fmt"
	"os"

	"strconv"
	"time"

	"github.com/slavikmanukyan/itm/fs"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/utils"
	"github.com/urfave/cli"
)

func Init(ctx *cli.Context, config itmconfig.ITMConfig) error {
	source := ctx.Parent().String("s")
	destinations, err := utils.ReadLines(filepath.Join(source, ".itm/destinations.itmi"))
	if ctx.Bool("f") != true {
		if err == nil && utils.Contains(destinations, config.DESTINATION) {
			return errors.New("Destination alerady inited (use -f to force init)")
		}
	}
	fmt.Println("Starting backup...")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	now := strconv.FormatInt(time.Now().Unix(), 10)
	if config.USE_SSH {
		_, err := fsftp.Client.Stat(config.DESTINATION)
		if err != nil {
			fsftp.Client.Mkdir(config.DESTINATION)
		}
		if ctx.Bool("f") {
			fsftp.Client.RemoveDirectory(filepath.Join(config.DESTINATION, ".itm"))
		}

		fsftp.Client.Mkdir(filepath.Join(config.DESTINATION, ".itm"))
		fsftp.Client.Mkdir(filepath.Join(config.DESTINATION, ".itm/files"))
		fsftp.Client.Mkdir(filepath.Join(config.DESTINATION, ".itm", "files", now))

		err = fsftp.CopyDirTo(source, config.DESTINATION, config, func(file string) {
			utils.ClearLine(80)
			fmt.Print("\rCopying file: ", file)
			if err != nil {
				return
			}
			hash.SaveFileHash(file, config, timestamp)
		})
		if err != nil {
			return err
		}
	} else {
		if ctx.Bool("f") {
			os.RemoveAll(filepath.Join(config.DESTINATION, ".itm"))
		}

		os.MkdirAll(filepath.Join(config.DESTINATION, ".itm", "files", now), os.ModePerm)
		err := fs.CopyDir(source, config.DESTINATION, config, now)
		if err != nil {
			return err
		}
	}
	_ = os.Mkdir(filepath.Join(source, ".itm"), os.ModePerm)
	if !utils.Contains(destinations, config.DESTINATION) {
		utils.WriteLines([]string{config.DESTINATION}, filepath.Join(source, ".itm/destinations.itmi"))
	}
	fmt.Println("\nBackup completed!")

	return nil
}
