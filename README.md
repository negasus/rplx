# RPLX [![Go Report Card](https://goreportcard.com/badge/github.com/negasus/rplx)](https://goreportcard.com/report/github.com/negasus/rplx) ![](https://github.com/negasus/rplx/workflows/Test/badge.svg)

> Pronounced as Replix

Golang library for multi master replication integer variables with TTL support

> todo

## Usage examples

```
func main() {
	r = rplx.New(
		rplx.WithRemoteNodesProvider(remoteNodes()),
	)

	ln, err := net.Listen("tcp4", "127.0.0.1:3001")  

	if err != nil {
		panic(err)
	}

	if err := r.StartReplicationServer(ln); err != nil {
		panic(err)
	}
}

func remoteNodes() []*rplx.RemoteNodeOption {
    nodes := make([]*rplx.RemoteNodeOption, 0)

    nodes = append(nodes, &rplx.DefaultRemoteNodeOption("127.0.0.1:3002")) 

    return nodes
}
```

Also see example in `test` folder

## Public API

### Get
> `Get(name string) (int64, error)`

Returns variable value or error, if variable expired or not exists

Errors:
- ErrVariableNotExists
- ErrVariableExpired

### Delete
> `Delete(name string) error`

Delete variable

Errors:
- ErrVariableNotExists

By fact this method sets TTL for variable to `Now - second`, send this info to replication and remove from local cache

### UpdateTTL

> `UpdateTTL(name string, ttl time.Time) error`

Update TTL for variable

Errors:
- ErrVariableNotExists

### Upsert

> `Upsert(name string, delta int64)`

Update variable value on provided delta, or create new variable, if not exists

### All

> `All() (notExpired map[string]int64, expired map[string]int64)`

Returns two maps, where variable name as map item key and variable value as map item value

First return param contains not expires variables. 
Second param contains expired (while not garbage colleced) variables

## Run integration tests

```
docker-compose up -d
docker build -t client -f ./test/client/Dockerfile .
docker run --rm --net host client
docker-compose down -v
```

### Additional 

- README.RUS.md