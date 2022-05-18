package utils

import (
	"os/exec"
)

func RunVideoPlayer(config VideoPlayerConfig) error {
	videoPlayerPath, err := exec.LookPath("vlc")
	if err != nil {
		videoPlayerPath = `/Applications/VLC.app/Contents/MacOS/VLC`
	}

	err = exec.Command(
		videoPlayerPath,
		getVlcArgs(config)...,
	).Run()

	if err != nil {
		if _, isExitError := err.(*exec.ExitError); !isExitError {
			return err
		}
	}

	return nil
}
