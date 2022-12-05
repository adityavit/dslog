# dslog

Book: Distributed Services with Go

Distributed Logging service

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

