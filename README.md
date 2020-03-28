Raft
====

### Build
```
go build client.go
```
```
go build server.go
```

### Test
```
go test
```

### Run
```
./server localhost:8001 localhost:8002 localhost:8003 localhost:8004 localhost:8005
./server localhost:8002 localhost:8001 localhost:8003 localhost:8004 localhost:8005
./server localhost:8003 localhost:8001 localhost:8002 localhost:8004 localhost:8005
./server localhost:8004 localhost:8001 localhost:8002 localhost:8003 localhost:8005
./server localhost:8005 localhost:8001 localhost:8002 localhost:8003 localhost:8004
```

```
./client localhost:8001 localhost:8002 localhost:8003 localhost:8004 localhost:8005
```

License
=======