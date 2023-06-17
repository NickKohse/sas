package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

//TODO most of thi file will need to have the functoins converted the return errors if they hit them instead of just printing something

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

func (m metadata) saveMetadata(filepath string) {
	bytes, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
	}
	writeErr := os.WriteFile(filepath, bytes, 0600)
	if writeErr != nil {
		fmt.Println(writeErr)
	}
}

func readMetadata(artifactPath string) *metadata {
	bytes, err := os.ReadFile("./repository_metadata/" + artifactPath + ".metadata")
	if err != nil {
		fmt.Println(err)
	}
	m := metadata{}
	unmarshalErr := json.Unmarshal(bytes, &m)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr)
	}
	return &m
}

// Same as above except instead of reading and unmarshaling the json into a struct, just return the json byte array
func readMetadataJson(artifactPath string) []byte {
	bytes, err := os.ReadFile("./repository_metadata/" + artifactPath + ".metadata")
	if err != nil {
		fmt.Println(err)
	}
	return bytes
}

func removeMetadata(artifactPath string) {
	err := os.Remove("./repository_metadata/" + artifactPath + ".metadata")
	if err != nil {
		fmt.Println(err)
	}
}
