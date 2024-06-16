package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func handleServerError(err error, w http.ResponseWriter) {
	fmt.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
}

func preFormCheck(w http.ResponseWriter, r *http.Request) (string, error) {
	firstSlash := strings.Index(r.URL.Path, "/")
	afterFirstSlash := r.URL.Path[firstSlash+1:]
	secondSlash := strings.Index(afterFirstSlash, "/")
	artifactName := afterFirstSlash[secondSlash+1:] //This converts something like /metadata/folder/testfile to folder/testfile

	if len(artifactName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No artifact specified in path"))
		return "", errors.New("no artifact found in path")
	}
	if !fileExists("./repository/" + artifactName) {
		w.WriteHeader(http.StatusNotFound)
		return "", errors.New("file not found in repository")
	}
	return artifactName, nil
}

func streamFile(sourceFile io.Reader, destFile io.Writer, w http.ResponseWriter) error {
	buffer := make([]byte, 4096)

	for { //TODO likely better to move the hash calculation into here, then we can only read the file once
		bytesRead, err := sourceFile.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			handleServerError(err, w)
			return errors.New("error reading file")
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

func checkFilesForMetadata(path string) {
	fmt.Println("Checking to ensure all files have metadata...")
	files, err := ioutil.ReadDir("repository" + path)
	if err != nil {
		fmt.Println("Error while preforming checkFilesForMetadata")
		fmt.Println(err)
	}
	for _, file := range files {
		if file.IsDir() {
			checkFilesForMetadata(filepath.Join(path, file.Name()))
		} else {
			if !fileExists("./.repository_metadata/" + filepath.Join(path, file.Name()) + ".metadata") {
				fmt.Printf("Warning: No Metadata found for %s. Generating it.\n", file.Name())
				metadataPath := "./.repository_metadata/" + path //TODO, move this logic to metadata file
				os.MkdirAll(metadataPath, os.ModePerm)           // This should be in some setup function
				go generateAndSaveMetadata(path+file.Name(), metadataCache, metadataQueue)
			}
		}
	}
	fmt.Println("Checking to ensure all files have metadata...Done.")
}

func checkForOrphanMetadata(path string) {
	fmt.Println("Checking to ensure each metadata file is associated with a file...")
	files, err := ioutil.ReadDir(".repository_metadata" + path)
	if err != nil {
		fmt.Println("Error while preforming checkForOrphanMetadata.")
		fmt.Println(err)
	}
	for _, file := range files {
		if file.IsDir() {
			checkForOrphanMetadata(filepath.Join(path, file.Name()))
		} else {
			if strings.HasSuffix(file.Name(), ".metadata") {
				if !fileExists("./repository/" + filepath.Join(path, strings.TrimSuffix(file.Name(), ".metadata"))) {
					fmt.Printf("Warning: No file associated with metadata %s. Removing it.\n", file.Name())
					os.Remove("./.repository_metadata/" + filepath.Join(path, file.Name()))
				}
			} else {
				fmt.Printf("Warning: file %s is not suffixed '.metadata', removing it...\n", file.Name())
				os.Remove("./.repository_metadata/" + filepath.Join(path, file.Name()))
			}
		}
	}
	fmt.Println("Checking to ensure each metadata file is associated with a file...Done.")
}
