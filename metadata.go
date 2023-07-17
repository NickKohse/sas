package main

import (
	"encoding/json"
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

func readMetadata(artifactPath string) (*metadata, error) {
	bytes, err := os.ReadFile("./repository_metadata/" + artifactPath + ".metadata")
	if err != nil {
		return nil, err
	}
	m := metadata{}
	unmarshalErr := json.Unmarshal(bytes, &m)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &m, nil
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
				queue[keys[m]].saveMetadata(keys[m]) //Ignoring erros here. fix that
				delete(queue, keys[m])
			}
		}
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
}
