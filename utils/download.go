package utils

import (
	"io"
	"net/http"
	"os"
)

func Download(url string, destination string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	file, err := os.Create(destination)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, res.Body); err != nil {
		return "", err
	}

	return destination, nil
}
