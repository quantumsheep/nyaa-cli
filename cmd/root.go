package cmd

import (
	"strings"

	"github.com/quantumsheep/nyaa-cli/ui"
	"github.com/urfave/cli/v2"
)

var allowedVideoPlayers = []string{
	"vlc",
	"airplay",
	"mplayer",
	"smplayer",
	"mpchc",
	"potplayer",
	"mpv",
	"omx",
	"webplay",
	"jack",
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
		&cli.BoolFlag{
			Name:  "peerflix",
			Usage: "run peerflix on the torrents",
		},
		&cli.StringFlag{
			Name:  "player",
			Usage: "video player used by peerflix. available options: " + strings.Join(allowedVideoPlayers, ", "),
			Value: "vlc",
		},
	},
	Action: func(c *cli.Context) error {
		return ui.NewUI(&ui.UIOptions{
			UsePeerflix:         c.Bool("peerflix"),
			PeerflixFullscreen:  c.Bool("fullscreen"),
			PeerflixVideoPlayer: c.String("player"),
			OutputDirectory:     c.String("dir"),
		}).Run()
	},
}
