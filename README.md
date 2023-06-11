# sas
Simple Artifact Store

## Basic Structure
- GET `/artifact/<path to artifact>`
- POST `/artifact/<path to artifact>`
- DELETE `/artifact/<path to artifact>`
- GET `/checksum/<path to artifact>`
- GET `/metadata/<path to artifact>`
- GET `/health`

## Example curl commands

### Storing a file
`curl -i -X POST -H "Content-Type: multipart/form-data" -F "artifact=@test" localhost:1997/artifact`

### Retrieving a file
`curl -X GET -F "artifact=test" localhost:1997/artifact -o output.txt`

## TODOs
- Send 5xx on any error encountered
- Stream files so it works for large files
- Make url paths wildcards so you can specify files to get in the url instead of form
