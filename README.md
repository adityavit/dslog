# dslog

Book: Distributed Services with Go

Distributed Logging service

# Chapter - 1

* Initializing go project

```bash
go mod init github.com/adityavit/dslog
```
* Adding a go dependency

```bash
go get -u github.com/gorilla/mux
go mod vendor
go mod tidy 
```

* Run the server:
```bash
make start
```

* Executing commands to store and extract information in server:

```bash
> echo "testing string" | base64
dGVzdGluZyBzdHJpbmcK
> curl -X POST localhost:8080 -d \
    '{"record": {"value": "dGVzdGluZyBzdHJpbmcK"}}'
{"offset":0}
> curl -sX GET localhost:8080 -d '{"offset": 0}' | jq -r ".record.value" | base64 -d
testing string
```

# Chapter - 2

* Install protobuf compiler 

1. Download the binary from ![latest release](https://github.com/protocolbuffers/protobuf/releases/download/v21.10/protoc-21.10-osx-x86_64.zip)
2. Copy the file to `/usr/local/protobuf`
3. Create link to the binary in `/usr/local/bin`

```bash
mkdir -p /usr/local/protobuf
mv ~/Downloads/protoc-21.10-osx-x86_64/* /usr/local/protobuf
ln -s /usr/local/protobuf/bin/protoc /usr/local/bin/protoc
```

* Generate protobuf structures

```bash
make generate
```

* Adding protobuf dependency

```bash
go get -u google.golang.org/protobuf
go mod vendor
go mod tidy 
```

