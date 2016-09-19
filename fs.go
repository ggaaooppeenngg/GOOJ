package main

import (
	"io"
	"io/ioutil"
	"os"
)

func SaveFile(key string, content string) error {
	return ioutil.WriteFile(key, []byte(content), 0644)
}

// GetFIle gets a file, it is the callers reponsibility to close file.
func GetFile(key string) (io.ReadCloser, error) {
	f, err := os.Open(key)
	if err != nil {
		return nil, err
	}
	return f, nil
}
