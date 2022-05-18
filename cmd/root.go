package cmd

import (
	"strings"

	"github.com/quantumsheep/nyaa-cli/ui"
	"github.com/urfave/cli/v2"
)

var allowedVideoPlayers = []string{
	"vlc",
}

var RootCmd = &cli.App{
	Name:  "nyaa",
	Usage: "Use nyaa.si from the CLI",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "dir",
			Aliases: []string{"d"},
			Usage:   "directory used to store the torrents",
			Value:   ".",
		},
		&cli.BoolFlag{
			Name:  "fullscreen",
			Usage: "run peerflix in fullscreen mode",
		},
		&cli.StringFlag{
			Name:  "player",
			Usage: "video player to run the videos on. available options: " + strings.Join(allowedVideoPlayers, ", "),
			Value: "vlc",
		},
	},
	Action: func(c *cli.Context) error {
		return ui.NewUI(&ui.UIOptions{
			VideoPlayer:     c.String("player"),
			Fullscreen:      c.Bool("fullscreen"),
			OutputDirectory: c.String("dir"),
		}).Run()
	},
}
