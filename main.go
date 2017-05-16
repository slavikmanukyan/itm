package main

import (
	"os"

	"path/filepath"

	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/urfave/cli"
)

func logError(err string) {

}

func main() {
	// source := flag.String("source", "", "source directory")
	// destination := flag.String("destination", "", "backup destination")
	// configSource := flag.String("config", ".itmconfig", "Config file destination")
	// // recover = := flag.Bool("recover", false, "recover files")

	// flag.Parse()
	var configSource string
	var config itmconfig.ITMConfig

	// if config.USE_SSH {
	// 	fsftp.InitClient(config)
	// 	err := fsftp.CopyFileFrom("/slavik/.gitignore", ".gitignore1")
	// 	fmt.Println(err)
	// }

	// if (*source != "" && *destination == "") || (*source == "" && *destination != "") {
	// 	logError("Wrong args")
	// 	return
	// }

	// if *destination != "" {
	// 	fs.CopyDir(*source, *destination)
	// }
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Value:       ".itmconfig",
			Destination: &configSource,
		},
		cli.StringFlag{
			Name:  "source, s",
			Value: ".",
			Usage: "File System source",
		},
		cli.StringFlag{
			Name:  "destination, d",
			Usage: "File System destination",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name: "init",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "force init",
				},
			},
			Action: func(ctx *cli.Context) error {
				if len(config.DESTINATION) == 0 {
					return cli.NewExitError("required destination", 1)
				}
				err := Init(ctx, config)
				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				return nil
			},
		},
	}

	app.Before = func(ctx *cli.Context) error {
		config = itmconfig.Parse(ctx.String("config"))
		if len(ctx.String("d")) > 0 {
			config.DESTINATION = ctx.String("d")
		}
		config.SOURCE = ctx.String("s")
		if config.IGNORE == nil {
			config.IGNORE = make(map[string]bool)
		}

		tmpMap := make(map[string]bool)
		config.IGNORE[".itm"] = true
		for key, value := range config.IGNORE {
			abs, _ := filepath.Abs(filepath.Join(config.SOURCE, key))
			tmpMap[abs] = value
		}
		config.IGNORE = tmpMap

		if config.USE_SSH {
			fsftp.InitClient(config)
		}
		return nil
	}

	app.Run(os.Args)
}
