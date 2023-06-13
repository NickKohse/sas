package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type metadata struct {
	CreateTime int64
	ModifyTime int64
	AccessTime int64
}

func newMetadata(ct int64, mt int64, at int64) *metadata {
	m := metadata{CreateTime: ct, ModifyTime: mt, AccessTime: at}
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

// func readMetadata(artifactPath string) *metadata {
// TODO read from the file, which will be repositopry_metadata/<artifact path>.metadata
// Then unmarsahl is and return pointer to metadataobject
// }
