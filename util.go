package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func handleServerError(err error, w http.ResponseWriter) {
	fmt.Println(err) //TODO log this to a file as well
	w.WriteHeader(http.StatusInternalServerError)
}

func fileCountAndSize(path string) (int, int64, error) {
	nf := 0
	var sf int64 = 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return -1, -1, err
	}
	for _, file := range files {
		if file.IsDir() {
			count, size, err := fileCountAndSize(filepath.Join(path, file.Name()))
			if err != nil {
				return -1, -1, err
			}
			nf += count
			sf += size
		} else {
			nf++
			sf += file.Size()
		}
	}
	return nf, sf, nil
}
