package utils

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os/exec"
)

func RunVideoPlayer(config VideoPlayerConfig) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\VideoLAN\VLC`, registry.QUERY_VALUE)
	if err != nil {
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Software\WOW6432Node\VideoLAN\VLC`, registry.SET_VALUE)
		if err != nil {
			return err
		}
	}
	defer k.Close()

	installDir, _, err := k.GetStringValue("InstallDir")
	if err != nil {
		return err
	}

	err = exec.Command(
		fmt.Sprintf("%s\\vlc.exe", installDir),
		getVlcArgs(config)...,
	).Run()

	if err != nil {
		if _, isExitError := err.(*exec.ExitError); !isExitError {
			return err
		}
	}

	return nil
}
