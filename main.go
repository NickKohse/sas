package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type response struct {
	Response any
	Location string
}

func sendFile(w http.ResponseWriter, r *http.Request) {
	artifactName, e := preFormCheck(w, r)
	if e != nil {
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	file, err := os.Open("./repository/" + artifactName)
	if err != nil {
		handleServerError(err, w)
		return
	}

	streamFile(file, w, w)

	file.Close()

	m, metadataErr := readMetadata(artifactName)
	if metadataErr != nil {
		handleServerError(metadataErr, w)
		return
	}

	m.AccessTime = time.Now().Unix()
	m.AccessCount++
	metadataQueue[artifactName] = m

	health.DownloadHits++
}

func sendMetadata(w http.ResponseWriter, r *http.Request) {
	artifactName, e := preFormCheck(w, r)
	if e != nil {
		return
	}

	m, metadataErr := readMetadata(artifactName)
	if metadataErr != nil {
		handleServerError(metadataErr, w)
		return
	}

	metadataJson, marshalErr := json.Marshal(m)
	if marshalErr != nil {
		handleServerError(marshalErr, w)
	}

	w.Write(metadataJson)
	health.MetadataHits++
}

func sendChecksum(w http.ResponseWriter, r *http.Request) {
	artifactName, e := preFormCheck(w, r)
	if e != nil {
		return
	}

	m, metadataErr := readMetadata(artifactName)
	if metadataErr != nil {
		handleServerError(metadataErr, w)
		return
	}

	w.Write([]byte(m.Sha256))
	health.DownloadHits++
}

func sendHealth(w http.ResponseWriter) {
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
	// This function is too long

	// 1024 MB limit in file size, should be configurable TODO
	r.ParseMultipartForm(1024 << 20)

	file, handler, err := r.FormFile("artifact") //TODO, can this be a wildcard too?
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unable to process artifact, it's possible the key doesnt exist in the form, or it was not a valid file.\n"))
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	path := strings.Replace(r.URL.Path, "/artifact", "", 1) // for now assume they wont specify the filename in the post path
	filePath := "./repository/" + path                      // TODO, eventually reponame will be specified in the url

	os.MkdirAll(filePath, os.ModePerm)
	destFile, destErr := os.Create(filePath + handler.Filename)
	if destErr != nil {
		handleServerError(err, w)
		return
	}

	streamFile(file, destFile, w)

	destFile.Close()
	file.Close()

	metadataPath := "./.repository_metadata" + path //TODO, move this logic to metadata file
	os.MkdirAll(metadataPath, os.ModePerm)          // This should be in some setup function
	go generateAndSaveMetadata(path+handler.Filename, metadataCache, metadataQueue)

	res := response{
		Response: "Successfully Uploaded File",
		Location: path + handler.Filename,
	}
	response, err := json.Marshal(res)
	if err != nil {
		handleServerError(err, w)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)
	health.UploadHits++
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	// Remove the file and its meta data, if we have any in memory references to that file remove them too
	artifactName, e := preFormCheck(w, r)
	if e != nil {
		return
	}

	err := os.Remove("./repository/" + artifactName)
	if err != nil {
		handleServerError(err, w)
		return
	} else {
		res := response{
			Response: "Successfully Deleted File",
			Location: artifactName,
		}
		response, err := json.Marshal(res)
		if err != nil {
			handleServerError(err, w)
		}
		w.Write(response)
	}
	delete(metadataQueue, artifactName)
	delete(metadataCache, artifactName)
	removeMetadata(artifactName)
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
		res := response{
			Response: "No results found",
			Location: "/", // Can only search top level right now, change this later
		}
		response, err := json.Marshal(res)
		if err != nil {
			handleServerError(err, w)
		}
		w.Write(response)
		return
	}
	res := response{
		Response: results,
		Location: "/", // Can only search top level right now, change this later
	}
	response, err := json.Marshal(res)
	if err != nil {
		handleServerError(err, w)
	}
	w.Write(response) //TODO: Add pagination
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
		sendHealth(w)
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
	go queueWriter(metadataQueue, 5)

	fmt.Println("Initiating startup checks...")
	go checkFilesForMetadata("")
	go checkForOrphanMetadata("")

	mux := http.NewServeMux()

	// ROUTES
	mux.HandleFunc("/artifact/", artifactHandler) // TODO LATER make this wildcard, as it currently only matches for exactly artifact
	mux.HandleFunc("/metadata/", metadataHandler)
	mux.HandleFunc("/checksum/", checksumHandler)
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/health", healthHandler)

	http.ListenAndServe(":1997", mux)
	fmt.Println("Running.")

}
