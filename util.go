package main

import (
	"errors"
	"fmt"
	"io"
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

func preCheck(w http.ResponseWriter, r *http.Request, ensureArtifactInReq bool, ensureFileInRepo bool) error {
	if r.FormValue("artifact") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No artifact key in form\n"))
		return errors.New("No artifact key in form\n")
	}
	if !fileExists("./repository/" + r.FormValue("artifact")) {
		w.WriteHeader(http.StatusNotFound)
		return errors.New("File not in repository")
	}
	return nil
}

func streamFile(sourceFile io.Reader, destFile io.Writer, w http.ResponseWriter) error {
	buffer := make([]byte, 4096)

	for { //TODO likely better to move the hash calculation into here, then we can only read the file once
		bytesRead, err := sourceFile.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			handleServerError(err, w)
			return errors.New("Error reading file.")
		}

		destFile.Write(buffer[:bytesRead])
	}
	return nil
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
