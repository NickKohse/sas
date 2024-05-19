package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type metadata struct {
	CreateTime  int64
	ModifyTime  int64
	AccessTime  int64
	Sha256      string
	Size        int64
	AccessCount int
}

// Like the normal constructor except assume create and modify times are now, access time is never (-1) and access count is 0
func newImplicitMetadata(sha string, size int64) *metadata {
	m := metadata{
		CreateTime:  time.Now().Unix(),
		ModifyTime:  time.Now().Unix(),
		AccessTime:  -1,
		Sha256:      sha,
		Size:        size,
		AccessCount: 0}
	return &m
}

func newMetadata(ct int64, mt int64, at int64, sha string, size int64, accessCount int) *metadata {
	m := metadata{
		CreateTime:  ct,
		ModifyTime:  mt,
		AccessTime:  at,
		Sha256:      sha,
		Size:        size,
		AccessCount: accessCount}
	return &m
}

// filepath refers to the in structure file path, i.e. not including system prefixes like repository_metadata or system suffixes like .metadata
func (m metadata) saveMetadata(filepath string) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	writeErr := os.WriteFile("./repository_metadata/"+filepath+".metadata", bytes, 0600)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

// Target is a file whose metadata we want to generate then save
func generateAndSaveMetadata(targetFile string, cache map[string]*metadata, queue map[string]*metadata) error { // We run this in a thread so it needs to print its own errors not return them
	update := false
	if fileExists("./repository_metadata/" + targetFile + ".metadata") {
		update = true
	}

	target, err := os.Open("./repository/" + targetFile)
	if err != nil {
		return err
	}
	defer target.Close()

	hasher := sha256.New()
	buffer := make([]byte, 4096)
	fileSize := 0

	for {
		bytesRead, err := target.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		hasher.Write(buffer[:bytesRead])
		fileSize += bytesRead
	}

	closeErr := target.Close()

	if closeErr != nil {
		return err
	}

	hashString := hex.EncodeToString(hasher.Sum(nil))
	m := &metadata{}
	if update {
		m, metadataErr := readMetadata(targetFile)
		if metadataErr != nil {
			return metadataErr
		}
		m.ModifyTime = time.Now().Unix()
		m.Size = int64(fileSize)
		m.Sha256 = hashString
		queue[targetFile] = m
		cache[targetFile] = m
		return nil
	} else {
		m = newImplicitMetadata(hashString, int64(fileSize))
		queue[targetFile] = m
		cache[targetFile] = m
		return nil
	}
}

func readMetadata(artifactPath string) (*metadata, error) {
	var m *metadata
	var ok bool
	m, ok = metadataCache[artifactPath]
	if !ok {
		bytes, err := os.ReadFile("./repository_metadata/" + artifactPath + ".metadata")
		if err != nil {
			return nil, err
		}

		unmarshalErr := json.Unmarshal(bytes, &m)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
	}

	return m, nil
}

// Same as above except instead of reading and unmarshaling the json into a struct, just return the json byte array
func readMetadataJson(artifactPath string) ([]byte, error) {
	bytes, err := os.ReadFile("./repository_metadata/" + artifactPath + ".metadata")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func removeMetadata(artifactPath string) error {
	err := os.Remove("./repository_metadata/" + artifactPath + ".metadata")
	if err != nil {
		return err
	}
	return nil
}

func queueWriter(queue map[string]*metadata, sleepDuration int) {
	for true {
		if len(queue) > 0 {
			keys := make([]string, 0, len(queue))
			for k := range queue {
				keys = append(keys, k)
			}
			for m := range keys {
				err := queue[keys[m]].saveMetadata(keys[m])
				if err != nil {
					fmt.Printf("Failed to save metadata: %s", keys[m])
				}
				delete(queue, keys[m])
			}
		}
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
}
