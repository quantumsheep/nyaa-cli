package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
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
			return NewUI(&UIOptions{
				UsePeerflix:        c.Bool("peerflix"),
				PeerflixFullscreen: c.Bool("fullscreen"),
				OutputDirectory:    c.String("dir"),
			}).Run()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
