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

## Config file
A config file can (must?) be specified. By default the program will attempt to read it from `config.toml` in the directory it's being run from, otherwise you can pass an argument to the program iwth the path of the file (todo).

### Exmaple config file
```toml
[general]
maxFileSizeMB = 2000000000 # todo
maxMemoryCacheMB = 100 # todo

[repos] # todo

[repos.imageStore]
name = "Image Store"
directory = "/data/images/sasImages"
maxRepoSizeGB = 30

[repos.releaseImages]
name = "Release Images"
directory = "/mnt/nfs/public/sasReleaseImages"
maxRepoSizeGB = 1024
```

### Key

#### General

##### maxFileSizeMB
The highest size in MB allowed for upload of a single file. Defaults to 1 GB.

##### maxMemoryCacheMB
The size cap for the metadata memory cache in MB , once hit, the least recently accessed items will be dropped from the cache.

#### repos
Multiple repos scan be defined, if none are defined a single repo, called 'repository' will be created in the directory the program is run from.

##### name
The name of the repo used for logging purposes. Required.

##### directory
The directory that will be used to store the files. If it doesn not exist the program will create it. If you specify `/data/repo` for example, and `/data/.repo_metadata` directory will also be created. Required.

##### maxRepoSzieGB
The maximum cumulative size of the files allowed in a repo. After the limit is hit furhter upload requests will be rejected. Optional, if not specified not limit will be enforced.


## TODOs
- divide health stat hits into 2xx/4xx/5xx
- Paginate search results
- should be able to search a sub directory
- Add config file
    - Including which port to run on, max file size allowed, etc.
- Once the config file is read, setup any necessary directories/anything else before the server starts running
- tests can parse the returned json and check the object has the right keys/reasonable values in the keys
- Limit metadata in memory array size
- Limit cumulative file sizes of a repo

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
- Auth
