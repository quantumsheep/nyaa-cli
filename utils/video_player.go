package utils

import "fmt"

type VideoPlayerConfig struct {
	VideoPlayer string
	Url         string
	Name        string
	OnTop       bool
	Fullscreen  bool
}

func getVlcArgs(config VideoPlayerConfig) []string {
	return []string{
		"-q",
		"--video-on-top",
		"--play-and-exit",
		fmt.Sprintf("--meta-title=\"%s\"", config.Name),
		config.Url,
	}
}
