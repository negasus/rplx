# RPLX
Golang library for multi master replication integer variables with TTL support



### Send replication data to remote nodes

Sending replication data to remote node occurs on:
- reached SyncInterval
- internal node buffer size exceeded limit (defaultMaxBufferSize)

## Usage example

- init clients

```
// Client1

logger, _ := zap.NewDevelopment()
grpcListener, _ = net.Listen("tcp4", "127.0.0.1:2001")

// create new rplx insctance
r1 := rplx.New("nodeID-1", logger)
go r1.ServeIncomingReplication(grpcListener)

// add client2 for replication 
r1.AddNode("nodeID-2", "127.0.0.1:2002", time.Second)
```

```
// Client2

logger, _ := zap.NewDevelopment()
grpcListener, _ = net.Listen("tcp4", "127.0.0.1:2002")

// create new rplx insctance
r2 := rplx.New("nodeID-2", logger)
go r2.ServeIncomingReplication(grpcListener)

// add client1 for replication 
r2.AddNode("nodeID-1", "127.0.0.1:2001", time.Second)
```

- change variable on client1
```
r1.Upsert("var1", 42)
```

- get variable on client2 after 1 second (sync interval)
```
value, _ := r2.get("var1") 

// 42
```

- update variable on client2
```
r2.Upsert("var1", -100)
```

- get variable on client1 after 1 second (sync interval)
```
value, _ := r1.get("var1") 

// -58
```

## API
### Upsert
// todo

### Get
// todo

### Delete
// todo

## ToDo

- garbage expired variable
- use sync.Pool for reusable structs
- if send to remote node failed, return variables to internal node buffer (?) Check for stamp for while resend to buffer