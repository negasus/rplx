package main

import (
	"fmt"
	"github.com/negasus/rplx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	r                  *rplx.Rplx
	defaultRplxAddr    = ":3000"
	defaultServiceAddr = ":2000"
)

func main() {

	rplxNodeName := os.Getenv("RPLX_NODE_NAME")
	rplxNodes := os.Getenv("RPLX_NODES")

	rplxAddr := os.Getenv("RPLX_ADDR")
	if rplxAddr == "" {
		rplxAddr = defaultRplxAddr
	}
	serviceAddr := os.Getenv("SERVICE_ADDR")

	if serviceAddr == "" {
		serviceAddr = defaultServiceAddr
	}

	rplxLogger, _ := zap.NewDevelopment()

	r = rplx.New(
		rplx.WithLogger(rplxLogger),
		rplx.WithNodeID(rplxNodeName),
		rplx.WithRemoteNodesProvider(remoteNodes(rplxNodes)),
		rplx.WithRemoteNodesCheckInterval(time.Second),
	)

	ln, err := net.Listen("tcp4", rplxAddr)

	if err != nil {
		panic(err)
	}

	go startMux(serviceAddr)

	if err := r.StartReplicationServer(ln); err != nil {
		panic(err)
	}
}

func remoteNodes(nodesStr string) rplx.RemoteNodesProvider {
	return func() []*rplx.RemoteNodeOption {
		nodes := make([]*rplx.RemoteNodeOption, 0)

		for _, node := range strings.Split(nodesStr, ",") {
			if node == "" {
				continue
			}
			nodes = append(nodes, &rplx.RemoteNodeOption{
				Addr:               node,
				DialOpts:           []grpc.DialOption{grpc.WithInsecure()},
				SyncInterval:       1,
				MaxBufferSize:      0,
				ConnectionInterval: 1,
				WaitSyncCount:      0,
			})
		}

		return nodes
	}
}

func startMux(addrMain string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", handlerGet)
	mux.HandleFunc("/upsert", handlerUpsert)
	mux.HandleFunc("/all", handlerAll)
	mux.HandleFunc("/delete", handlerDelete)
	mux.HandleFunc("/update-ttl", handlerUpdateTTL)

	if err := http.ListenAndServe(addrMain, mux); err != nil {
		log.Printf("error listen mux, %v", err)
		os.Exit(1)
	}
}

func fprint(w io.Writer, format string, a ...interface{}) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		log.Panicf("error write, %v", err)
	}
}

func handlerDelete(w http.ResponseWriter, req *http.Request) {
	varName := req.URL.Query().Get("name")
	if varName == "" {
		w.WriteHeader(400)
		fprint(w, "empty variable name")
		return
	}

	log.Printf("delete '%s'", varName)

	err := r.Delete(varName)

	if err != nil {
		fprint(w, "error delete variable '%s', %v", varName, err)
		return
	}

	fprint(w, "ok")
}

func handlerUpdateTTL(w http.ResponseWriter, req *http.Request) {
	varName := req.URL.Query().Get("name")
	if varName == "" {
		w.WriteHeader(400)
		fprint(w, "empty variable name")
		return
	}

	ttlStr := req.URL.Query().Get("ttl")
	ttlSeconds, err := strconv.Atoi(ttlStr)
	if err != nil {
		w.WriteHeader(400)
		fprint(w, "error convert ttl '%s' to int, %v", ttlStr, err)
		return

	}

	log.Printf("update ttl '%s' with '%d'", varName, ttlSeconds)

	err = r.UpdateTTL(varName, time.Now().Add(time.Duration(ttlSeconds)*time.Second))
	if err != nil {
		w.WriteHeader(400)
		fprint(w, "error update ttl for var '%s', %v", varName, err)
		return
	}

	fprint(w, "ok")
}

func handlerAll(w http.ResponseWriter, req *http.Request) {
	notExpired, expired := r.All()
	fprint(w, "[not expired]")
	for name, value := range notExpired {
		fprint(w, "%s: %d", name, value)
	}
	fprint(w, "[expired]")
	for name, value := range expired {
		fprint(w, "%s: %d", name, value)
	}
}

func handlerGet(w http.ResponseWriter, req *http.Request) {
	varName := req.URL.Query().Get("name")
	if varName == "" {
		w.WriteHeader(400)
		fprint(w, "empty variable name")
		return
	}

	log.Printf("get '%s'", varName)

	value, err := r.Get(varName)

	if err != nil {
		fprint(w, "error get variable '%s', %v", varName, err)
		return
	}

	fprint(w, "value: %d", value)
}

func handlerUpsert(w http.ResponseWriter, req *http.Request) {
	varName := req.URL.Query().Get("name")
	if varName == "" {
		w.WriteHeader(400)
		fprint(w, "empty variable name")
		return
	}

	valueStr := req.URL.Query().Get("value")
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		w.WriteHeader(400)
		fprint(w, "error convert value '%s' to int, %v", valueStr, err)
		return

	}

	log.Printf("upsert '%s' with %d", varName, value)

	newValue := r.Upsert(varName, int64(value))

	fprint(w, "new value: %d", newValue)
}
