package main

import (
	"os"

	"path/filepath"

	"fmt"

	"time"

	"github.com/olekukonko/ts"
	fsftp "github.com/slavikmanukyan/itm/fs/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/status"
	"github.com/urfave/cli"
)

func logError(err string) {

}

func main() {
	var configSource string
	var config itmconfig.ITMConfig
	timeLayout := "02-01-2006 15:04"
	timeLayout2 := "02-01-2006"

	app := cli.NewApp()

	app.Name = "itm"
	app.Usage = "Incremental Time Machine - incrementally backing up data"
	app.Version = "1.0.0"
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "time, t",
					Usage: "Diff time (12-01-2016 15:24, 10-10-2016)",
				},
			},
			Action: func(ctx *cli.Context) error {
				if len(config.DESTINATION) == 0 {
					return cli.NewExitError("required destination", 1)
				}
				var timestamp int64
				timestamp = 0
				if ctx.String("time") != "" {
					t, err := time.Parse(timeLayout, ctx.String("time"))
					if err != nil {
						t, err = time.Parse(timeLayout2, ctx.String("time"))
						if err != nil {
							return cli.NewExitError("Wrong time format", 1)
						}
					}
					timestamp = t.UTC().Unix()
				}
				added, deleted, changed, _ := status.GetStatus(config, timestamp)
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
			Name: "backup",
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
					t, err = time.Parse(timeLayout2, ctx.String("time"))
					if err != nil {
						return cli.NewExitError("Wrong time format", 1)
					}
				}
				Restore(config, t.UTC().Unix())
				return nil
			},
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if (ctx.NArg() == 0) || ctx.Args().Get(0) == "help" || ctx.Args().Get(0) == "h" {
			return nil
		}
		t := false
		for _, c := range ctx.App.Commands {
			if c.HasName(ctx.Args().Get(0)) {
				t = true
				break
			}
		}
		if !t {
			return nil
		}
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
		config.IGNORE[configSource] = true
		for key, value := range config.IGNORE {
			abs, err := filepath.Abs(filepath.Join(config.SOURCE, key))
			if err == nil {
				tmpMap[abs] = value
			}
		}
		config.IGNORE = tmpMap
		_, err := ts.GetSize()
		if err != nil {
			config.IS_TERMINAL = false
		} else {
			config.IS_TERMINAL = true
		}

		if config.USE_SSH {
			fsftp.InitClient(config)
		}
		return nil
	}

	app.Run(os.Args)
}
