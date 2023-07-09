package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func sendFile(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("artifact") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No artifact key found in form\n"))
		return
	}
	if !fileExists("./repository/" + r.FormValue("artifact")) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Send the file to the client
	fileBytes, err := ioutil.ReadFile("./repository/" + r.FormValue("artifact")) //TODO Stream this instead of reading it all into memory
	if err != nil {
		handleServerError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)

	m, metadataErr := readMetadata(r.FormValue("artifact"))
	if metadataErr != nil {
		handleServerError(metadataErr, w)
		return
	}
	m.AccessTime = time.Now().Unix()
	m.AccessCount++

	saveErr := m.saveMetadata(r.FormValue("artifact"))
	if saveErr != nil {
		handleServerError(saveErr, w)
		return
	}
}

func sendMetadata(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("artifact") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No artifact key found in form\n"))
		return
	}
	if !fileExists("./repository/" + r.FormValue("artifact")) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metadataJson, err := readMetadataJson(r.FormValue("artifact"))
	if err != nil {
		handleServerError(err, w)
		return
	}
	w.Write(metadataJson)
}

func sendChecksum(w http.ResponseWriter, r *http.Request) {
	// TODO eventually we will keep the checksum in the metadata and not calculate it here
	if r.FormValue("artifact") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No artifact key found in form\n"))
		return
	}
	if !fileExists("./repository/" + r.FormValue("artifact")) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	file, err := os.Open("./repository/" + r.FormValue("artifact"))
	if err != nil {
		handleServerError(err, w)
		return
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		handleServerError(err, w)
		return
	}
	hashString := hex.EncodeToString(hasher.Sum(nil)) + "\n"
	w.Write([]byte(hashString))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unable to process artifact, it's possible the key doesnt exist in the form, or it was not a valid file.\n"))
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
		handleServerError(err, w)
		return
	}

	path := strings.Replace(r.URL.Path, "artifact", "", 1) // for now assume they wont specify the filename in the post path
	filePath := "./repository" + path                      // TODO, eventually reponame will be specified in the url
	os.MkdirAll(filePath, os.ModePerm)

	update := false
	if fileExists(filePath + handler.Filename) {
		update = true
	}

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	// nick - here we will actaully need to create a file in the correct directory after reading the url
	writeErr := os.WriteFile(filePath+handler.Filename, fileBytes, 0600)
	if writeErr != nil {
		handleServerError(writeErr, w)
		return
	}

	metadataPath := "./repository_metadata" + path //TODO, move this logic to metadata file
	os.MkdirAll(metadataPath, os.ModePerm)
	//TODO move metadata saving to a thread so we can return faster. ALSO TODO, need to differentiate between a new upload and a re-upload
	hasher := sha256.New()
	_, hashErr := hasher.Write(fileBytes)
	if hashErr != nil {
		handleServerError(hashErr, w)
		return
	}
	hashString := hex.EncodeToString(hasher.Sum(nil))
	m := &metadata{}
	if update {
		m, metadataErr := readMetadata(handler.Filename)
		if metadataErr != nil {
			handleServerError(metadataErr, w)
			return
		}
		m.ModifyTime = time.Now().Unix()
		m.Size = handler.Size
		m.Sha256 = hashString
	} else {
		m = newImplicitMetadata(hashString, handler.Size)
	}

	saveErr := m.saveMetadata(handler.Filename)
	if saveErr != nil {
		handleServerError(saveErr, w)
		return
	}

	// return that we have successfully uploaded our file!
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	// Remove the file and its meta data, if we have any in memory references to that file remove them too
	if r.FormValue("artifact") == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No artifact key found in form\n"))
		return
	}
	if !fileExists("./repository/" + r.FormValue("artifact")) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := os.Remove("./repository/" + r.FormValue("artifact"))
	if err != nil {
		handleServerError(err, w)
		return
	} else {
		w.Write([]byte("Successfully deleted file: " + r.FormValue("artifact") + "\n")) //print success if file is removed
	}
	removeMetadata(r.FormValue("artifact"))
}

func searchRepo(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var results []string
	entries, err := os.ReadDir("./repository")
	if err != nil {
		handleServerError(err, w)
		return
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), query) {
			results = append(results, e.Name())
		}
	}
	if len(results) == 0 {
		w.Write([]byte("No results found.\n"))
		return
	}
	w.Write([]byte(strings.Join(results, "\n") + "\n")) //TODO: Add pagination
}

func artifactHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sendFile(w, r)
	case "POST":
		recieveFile(w, r)
	case "DELETE":
		deleteFile(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed\n"))
	}
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sendMetadata(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed\n"))
	}
}

func checksumHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sendChecksum(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed\n"))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Return uptime, storage stats, files managed etc.
	switch r.Method {
	case "GET":
		sendHealth(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed\n"))
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		searchRepo(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed\n"))
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
	http.HandleFunc("/search", searchHandler)

	http.ListenAndServe(":1997", nil)

}
