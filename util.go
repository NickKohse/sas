package main

import (
	"fmt"
	"net/http"
	"os"
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
