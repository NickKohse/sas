# Test Cases

## Happy Path
- Upload a file
- Download that file (check its the same)
- Download the metadata (cehck the sha, size and timestamps are correct, as well as access count == 1)
- Use the checksum endpoint to again verify that its got the right value
- search for the file
- Remove the file
- Hit the download, metadata and checksum endpoints for it and ensure you get 404 for all of them

## Bad Path
- Hit the upload endpoint but dont specify a file, expect 400
- Hit the download, checksum, metadata endpoints, dont specify an artifact name, expect 400
- Hit download, checksum and metadata endponts for a file that never existed, expect 404
