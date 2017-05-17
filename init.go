package main

import (
	"errors"
	"path/filepath"

	"os"

	"strconv"
	"time"

	"github.com/slavikmanukyan/itm/fs"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
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
	if config.USE_SSH {
		if ctx.Bool("f") {
			fsftp.Client.RemoveDirectory(filepath.Join(config.DESTINATION, ".itm"))
		}
		fsftp.Client.Mkdir(filepath.Join(config.DESTINATION, ".itm"))
		fsftp.Client.Mkdir(filepath.Join(config.DESTINATION, ".itm/files"))
		err := fsftp.CopyDirTo(source, config.DESTINATION, config, strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			return err
		}
	} else {
		if ctx.Bool("f") {
			os.RemoveAll(filepath.Join(config.DESTINATION, ".itm"))
		}
		now := strconv.FormatInt(time.Now().Unix(), 10)
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

	return nil
}
