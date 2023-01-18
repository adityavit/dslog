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

# Chapter - 3

Creating log structure

* Record—the data stored in our log.
* Store—the file we store records in.
* Index—the file we store index entries in.
* Segment—the abstraction that ties a store and an index together. 
* Log—the abstraction that ties all the segments together.

Q&A:

* What is Big Endian/Little Endian Encoding?
  * BigEndian/Little Endian encoding is way to store the bytes in the memory or network.
  * In the Big Endian ordering the highest significant byte is store before the lower significant byte, whereas in the little endian this is opposite.
  * Example: A 4 byte word (uint32) let's say x0A0B0C0D = 168496141 will be stored as x0A, x0B, x0C, x0D in the memory in Big endian encoding from left to right. This more like natural order.
  * In the case of little Endian the same integer x0A0B0C0D = 168496141 will be stored as x0D, x0C, x0B, x0A in the memory addresses from left to right.  This is like writing number in opposite.
  * In the x86_64 or arm processors memory layout is done in little endian encoding.
  * In the network transfer the big endian encoding is used, also called as network byte order.
  * More Info: [Endian Wiki](https://en.wikipedia.org/wiki/Endianness), [Another explanation article](https://www.section.io/engineering-education/what-is-little-endian-and-big-endian/)

Code for encoding 32 bit (4 byte) unsigned int in little endian and big endian encoding in go and storing it in a byte array shows difference clearly.

[`little Endian`](https://go.dev/src/encoding/binary/binary.go#L84) : As can be seen the least significant byte is stored in the first byte, following the next 3 bytes 

```go
func (littleEndian) PutUint32(b []byte, v uint32) { // v => x0A0B0C0D
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v) // b[0] = 0D
	b[1] = byte(v >> 8) // b[1] = 0C
	b[2] = byte(v >> 16) // b[2] = 0B
	b[3] = byte(v >> 24) // b[3] = 0A 
}
```

[`big Endian`](https://go.dev/src/encoding/binary/binary.go#L161) : As can be seen the highest significant byte is stored in the first byte by first rightsizing 24 bits to right, following the next 3 bytes

```go
func (bigEndian) PutUint32(b []byte, v uint32) {  // v => x0A0B0C0D
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 24) // b[0] = 0A
	b[1] = byte(v >> 16) // b[1] = 0B
	b[2] = byte(v >> 8) // b[2] = 0C
	b[3] = byte(v) // b[3] = 0D
}
```

* How to use the `binary` package to encode different values into bytes?
* How to use the `bufio` package to write and read data from the file using buffer?
* What is mmap file? How mmap is used, what are the common API's?
* Why index file is truncated before loading in the memory as mmap? 
* Why offset in the index are used as relative offset? Or why the offset is stored as 4 bytes? 
* What is the difference between `sync.Mutex` and `sync.RWMutex`? When to use RWMutex over Mutex?
* Implement the io.Reader and io.Writer interface in golang on a string buffer.

# Chapter 4

* What are some of the challenges when building distributed service APIs?
  * The main two challenges which come across when designing distributed APIs' are
    * Maintaining compatibility between different versions of the API's between client and server. In gRPC this is done by versioning the API. 
      So that older clients can still work with the API's.
    * Maintaining performance between the server and client. At the API level gRPC reduce overhead in marshalling and unmarshalling and connection time between server and client.
* 




