package cmd

import (
	"github.com/quantumsheep/nyaa-cli/ui"
	"github.com/urfave/cli/v2"
)

var RootCmd = &cli.App{
	Name:  "nyaa",
	Usage: "Use nyaa.si from the CLI",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "dir",
			Aliases: []string{"d"},
			Value:   ".",
			Usage:   "directory used to store the torrents",
		},
		&cli.BoolFlag{
			Name:  "peerflix",
			Usage: "run peerflix on the torrents",
		},
		&cli.BoolFlag{
			Name:  "fullscreen",
			Usage: "run peerflix in fullscreen mode",
		},
	},
	Action: func(c *cli.Context) error {
		return ui.NewUI(&ui.UIOptions{
			UsePeerflix:        c.Bool("peerflix"),
			PeerflixFullscreen: c.Bool("fullscreen"),
			OutputDirectory:    c.String("dir"),
		}).Run()
	},
}
