package main

import (
	"os"

	"path/filepath"

	"fmt"

	"time"

	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/status"
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
	timeLayout := "15-01-2006 15:04"

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
			Destination: &configSource,
			Usage:       "Config file destination",
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
		cli.Command{
			Name: "status",
			Action: func(ctx *cli.Context) error {
				if len(config.DESTINATION) == 0 {
					return cli.NewExitError("required destination", 1)
				}
				added, deleted, changed, _ := status.GetStatus(config)
				if len(added) == 0 && len(changed) == 0 && len(deleted) == 0 {
					fmt.Println("Everything is up-to-date")
					return nil
				}
				if len(added) > 0 {
					fmt.Println("\nAdded files: ")
					for _, file := range added {
						fmt.Println("            ", file)
					}
				}
				if len(deleted) > 0 {
					fmt.Println("\nDeleted files: ")
					for _, file := range deleted {
						fmt.Println("              ", file)
					}
				}
				if len(changed) > 0 {
					fmt.Println("\nChanged files: ")
					for _, file := range changed {
						fmt.Println("              ", file)
					}
				}
				return nil
			},
		},
		cli.Command{
			Name: "save",
			Action: func(ctx *cli.Context) error {
				if len(config.DESTINATION) == 0 {
					return cli.NewExitError("required destination", 1)
				}
				Save(config)
				return nil
			},
		},
		cli.Command{
			Name: "back",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "time, t",
					Value: time.Now().Format(timeLayout),
					Usage: "Backup time (12-01-2016 15:24, 10-10-2016)",
				},
			},
			Action: func(ctx *cli.Context) error {
				if len(config.DESTINATION) == 0 {
					return cli.NewExitError("required destination", 1)
				}
				t, err := time.Parse(timeLayout, ctx.String("time"))
				if err != nil {
					return cli.NewExitError("Wrong time format", 1)
				}
				Restore(config, t.UTC().Unix())
				return nil
			},
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if configSource == "" {
			configSource = filepath.Join(ctx.String("s"), ".itmconfig")
		}
		config = itmconfig.Parse(configSource)
		if len(ctx.String("d")) > 0 {
			config.DESTINATION = ctx.String("d")
		}
		config.SOURCE, _ = filepath.Abs(ctx.String("s"))
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
