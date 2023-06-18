# Design Doc

## Metadata
### What data
- Creation timestamp
- Last modified timestamp
- sha256
- Last access timestamp
- Access count
- Size

### Processes
#### A new file is uploaded
- Metadata is calculated and stored in memory cache, creation timestamp, sha256, and size are calculated modified timpstamp will be the same as creation, last access will be nil, access count 0
- Metadata is written to disk, after ther http response has been sent

#### A file is updated
- sha256 and size are recalculated, modified time is updated, other fields are left the same
- Metadata is written to disk, after ther http response has been sent

#### A file is retrived
- Last access and access count are updated
- Metadata is written to disk, after ther http response has been sent

#### A file is removed
- Metedata is deleted from disk after the file is removed
- Metadata is cleared from in memory cache


### Notes
- The in memory cache will always have the most recent data available, therefore the metadata endpoint will always respond by reading from memory unless its not in memory yet. we will load it if its not and keep it forever. If this becomes a scale issue there would either need to be a size or time limit for data stored im mem.
- Due to the time delay between the in-memory data getting updated and it being written to disk its possible for the data to become inaccurate if the server crashes at the wonr time. mainly this would affect sha256 and size other data might be lost as well, but this would be unrevocerable
    - Because of this there must be a way to force a metadata rebuild, or it could just happen automatically when it sees a file whose metadata mtime is older than the file mtime it refers to.
- Metadat will be stored in a parallel repository, for example if the user specifies a repo called buildFiles, the metadata will be stored in buildFiles_meta


## Health Stats
### What data
- Uptime
- \# Files managed
- Total space used (eventually space used per repo as well)
- Endpoint usage stats since last reboot