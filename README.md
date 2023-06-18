# sas
Simple Artifact Store

## Basic Structure
- GET `/artifact`
- POST `/artifact`
- DELETE `/artifact`
- GET `/checksum`
- GET `/metadata`
- GET `/health`
- GET `/search?q=<search term>`

## Example curl commands

### Storing/updating a file
`curl -i -X POST -H "Content-Type: multipart/form-data" -F "artifact=@test" localhost:1997/artifact`

### Retrieving a file
`curl -X GET -F "artifact=test" localhost:1997/artifact -o output.txt`

### Removing a file
`curl -X DELETE -F "artifact=test" localhost:1997/artifact`

### Retrieving Metadata
`curl -X GET -F "artifact=test" localhost:1997/metadata`

## TODOs
- Send 5xx on any error encountered
- Stream files so it works for large files
- Make url paths wildcards so you can specify files to get in the url instead of form
- Write simple test harness script
- Return 404 if we get a request for a file/metadata that doesnt exist
- Im memory metadata cache

## Rough Version Plan
### v1
Fully working, one folder/one repo structure thats searchable.

