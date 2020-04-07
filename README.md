Raft
====
An implementation of the [Raft](https://raft.github.io/raft.pdf) consensus algorithm in Go. The following implementation includes: Leader Election.
Log Replication, Membership Changes and Log Compaction is currently work in progress.

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