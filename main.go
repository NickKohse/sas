package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func sendFile(w http.ResponseWriter, r *http.Request) {
	// Send the file to the client
	fmt.Printf(r.FormValue("artifact"))
	fileBytes, err := ioutil.ReadFile("./repository/" + r.FormValue("artifact")) //TODO Stream this instead of reading it all into memory
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}

func sendMetadata() {
	// Send the contents of the meta data file in json
}

func sendChecksum() {
	// Send the sha256 checksum of the file

}

func sendHealth(w http.ResponseWriter, r *http.Request) {
	uptime := time.Now().Unix() - startTime

	w.Write([]byte("Uptime: " + fmt.Sprint(uptime) + "\n"))
}

func recieveFile(w http.ResponseWriter, r *http.Request) {
	// Write the file to the specified path, creating any folder necessary
	// Ideally snaitize path so nothing funky is going on.
	// TODO any place we have error handling we need to return 500

	// 1024 MB limit in file size, should be configurable TODO
	r.ParseMultipartForm(1024 << 20)
	// look for key 'artifact'
	file, handler, err := r.FormFile("artifact")
	if err != nil {
		// Write a message and send an appropriate error code
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// read all of the contents of our uploaded file into a
	// byte array
	// TODO would be better to do this by streaming
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	path := strings.Replace(r.URL.Path, "artifact", "", 1) // for now assume they wont specify the filename in the post path
	path = "./repository" + path                           // TODO, eventually reponame will be specified in the url
	os.MkdirAll(path, os.ModePerm)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	// nick - here we will actaully need to create a file in the correct directory after reading the url
	writeErr := os.WriteFile(path+handler.Filename, fileBytes, 0600)
	if writeErr != nil {
		fmt.Println(err)
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func deleteFile() {
	// Remove the file and its meta data, if we have any in memory references to that file remove them too
}

func artifactHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sendFile(w, r)
	case "POST":
		recieveFile(w, r)
	case "DELETE":
		w.Write([]byte("Received a DELETE request\n"))
		deleteFile()
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not Implemented\n"))
	}
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write([]byte("Received a GET request\n"))
		sendMetadata()
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not Implemented\n"))
	}
}

func checksumHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write([]byte("Received a GET request\n"))
		sendChecksum()
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not Implemented\n"))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Return uptime, storage stats, files managed etc.
	switch r.Method {
	case "GET":
		sendHealth(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("Not Implemented\n"))
	}
}

var startTime int64

func main() {
	fmt.Println("Starting SAS...")
	startTime = time.Now().Unix()
	fmt.Println("This is where it would read in the config file...")
	fmt.Println("Running.")

	// ROUTES
	http.HandleFunc("/artifact", artifactHandler) // TODO LATER make this wildcard, as it currently only matches for exactly artifact
	http.HandleFunc("/metadata", metadataHandler)
	http.HandleFunc("/checksum", checksumHandler)
	http.HandleFunc("/health", healthHandler)

	http.ListenAndServe(":1997", nil)

}
