# sas
Simple Artifact Store

## Basic Structure
- GET `/artifact/<name>`
- POST `/artifact`
- DELETE `/artifact/<name>`
- GET `/checksum/<name>`
- GET `/metadata/<name>`
- GET `/health`
- GET `/search?q=<search term>`

## Example curl commands

### Storing/updating a file
`curl -i -X POST -H "Content-Type: multipart/form-data" -F "artifact=@test" localhost:1997/artifact`

### Retrieving a file
`curl -X GET localhost:1997/artifact/test -o output.txt`

### Removing a file
`curl -X DELETE localhost:1997/artifact/test`

### Retrieving Metadata
`curl -X GET localhost:1997/metadata/test`

### Searching for a file 
`curl -X GET localhost:1997/search?q=test`

## TODOs
- Make url paths wildcards so you can specify files to get in the url instead of form
- divide health stat hits into 2xx/4xx/5xx
- Paginate search results
- should be able to search a sub directory
- Add config file
    - Including which port to run on, max file size allowed, etc.
- Once the config file is read, setup any necessary directories/anything else before the server starts running
- tests can parse the returned json and check the object has the right keys/reasonable values in the keys

## Rough Version Plan
### v0 beta
- one folder/one repo structure thats searchable.
- error handling for non existant files/server side errors

### v0.x beta
- In memory metadata - write it to disk in a seperate thread
- Flesh out stats
- Add file streaming
- Send checksum from metadata
- Always return JSON
- Startup Sanity checks
- Paginate search results
- Further validation on tests besides just output and reponse code

### v1
- Config file, can specify multiple repos to create 
- Make urls wildcard

### v1.x
- Paginate search results, add searching a sub-directory
- More in depth health stats
