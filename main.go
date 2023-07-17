package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	var m *metadata
	var ok bool

	m, ok = metadataCache[r.FormValue("artifact")]
	if !ok {
		var metadataErr error
		m, metadataErr = readMetadata(r.FormValue("artifact"))
		if metadataErr != nil {
			handleServerError(metadataErr, w)
			return
		}
	}

	m.AccessTime = time.Now().Unix()
	m.AccessCount++
	metadataQueue[r.FormValue("artifact")] = m

	health.DownloadHits++
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

	var metadataJson []byte

	metadata, ok := metadataCache[r.FormValue("artifact")]
	if ok {
		var marshalErr error
		metadataJson, marshalErr = json.Marshal(metadata)
		if marshalErr != nil {
			handleServerError(marshalErr, w)
		}
	}
	metadataJson, err := readMetadataJson(r.FormValue("artifact"))
	if err != nil {
		handleServerError(err, w)
		return
	}
	w.Write(metadataJson)
	health.MetadataHits++
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

	var m *metadata
	var ok bool

	m, ok = metadataCache[r.FormValue("artifact")]
	if !ok {
		var metadataErr error
		m, metadataErr = readMetadata(r.FormValue("artifact"))
		if metadataErr != nil {
			handleServerError(metadataErr, w)
			return
		}
	}
	w.Write([]byte(m.Sha256))
	health.DownloadHits++
}

func sendHealth(w http.ResponseWriter, r *http.Request) {
	response, err := buildHealthStats(health)
	if err != nil {
		handleServerError(err, w)
		return
	}

	w.Write(response)
	health.HealthHits++
}

func recieveFile(w http.ResponseWriter, r *http.Request) {
	// Write the file to the specified path, creating any folder necessary
	// Ideally snaitize path so nothing funky is going on.

	// 1024 MB limit in file size, should be configurable TODO
	r.ParseMultipartForm(1024 << 20)

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

	// TODO would be better to do this by streaming
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		handleServerError(err, w)
		return
	}

	path := strings.Replace(r.URL.Path, "/artifact", "", 1) // for now assume they wont specify the filename in the post path
	filePath := "./repository/" + path                      // TODO, eventually reponame will be specified in the url

	os.MkdirAll(filePath, os.ModePerm)

	update := false
	if fileExists(filePath + handler.Filename) {
		update = true
	}

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
	metadataCache[path+handler.Filename] = m
	metadataQueue[path+handler.Filename] = m

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Successfully Uploaded File\n")
	health.UploadHits++
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
	delete(metadataQueue, r.FormValue("artifact"))
	delete(metadataCache, r.FormValue("artifact"))
	removeMetadata(r.FormValue("artifact"))
	health.DeleteHits++
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
	health.SearchHits++
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

var health *healthStats
var metadataCache = make(map[string]*metadata)
var metadataQueue = make(map[string]*metadata)

func main() {
	fmt.Println("Starting SAS...")
	health = newHealthStats(time.Now().Unix())
	fmt.Println("This is where it would read in the config file...")
	fmt.Println("Starting metadata writer thread...")
	go queueWriter(metadataQueue, 5) // Make the time configurable
	fmt.Println("Running.")

	// ROUTES
	http.HandleFunc("/artifact", artifactHandler) // TODO LATER make this wildcard, as it currently only matches for exactly artifact
	http.HandleFunc("/metadata", metadataHandler)
	http.HandleFunc("/checksum", checksumHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/search", searchHandler)

	http.ListenAndServe(":1997", nil)

}
