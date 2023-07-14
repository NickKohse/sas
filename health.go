package main

import (
	"encoding/json"
	"time"
)

type healthStats struct {
	StartTime     int64
	Uptime        int64
	UploadHits    int
	DownloadHits  int
	MetadataHits  int
	ChecksumHits  int
	HealthHits    int
	SearchHits    int
	DeleteHits    int
	FilesManaged  int
	FileSizeBytes int64
}

// Called on server startup
func newHealthStats(startTime int64) *healthStats {
	h := healthStats{
		StartTime:     startTime,
		Uptime:        0,
		UploadHits:    0,
		DownloadHits:  0,
		MetadataHits:  0,
		ChecksumHits:  0,
		HealthHits:    0,
		SearchHits:    0,
		DeleteHits:    0,
		FilesManaged:  0,
		FileSizeBytes: 0,
	}
	return &h
}

// Called when sending response on the health endpoint
func buildHealthStats(h *healthStats) ([]byte, error) {
	h.Uptime = time.Now().Unix() - h.StartTime
	count, size, err := fileCountAndSize("repository")
	if err != nil {
		return nil, err
	}
	h.FilesManaged = count
	h.FileSizeBytes = size
	bytes, jsonErr := json.Marshal(h)
	if jsonErr != nil {
		return nil, err
	}

	return bytes, nil
}
