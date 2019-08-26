package main

import (
	"flag"
	"fmt"
	"github.com/negasus/rplx"
	"go.uber.org/zap"
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
	remoteNodeSyncInterval = time.Second * 60
	r                      *rplx.Rplx
)

func main() {

	addrMain := flag.String("addr-main", ":2000", "")
	addrRplx := flag.String("addr-rplx", ":2001", "")
	nodes := flag.String("nodes", "", "")
	nodeName := flag.String("name", "", "")
	flag.Parse()

	rplxLogger, _ := zap.NewDevelopment()

	r = rplx.New(rplx.WithLogger(rplxLogger), rplx.WithNodeMaxBufferSize(0), rplx.WithNodeID(*nodeName))

	ln, err := net.Listen("tcp4", *addrRplx)

	if err != nil {
		panic(err)
	}

	for _, node := range strings.Split(*nodes, ",") {
		log.Printf("add remote node %s", node)
		go func(node string) {
			//if err := r.addRemoteNode(node, remoteNodeSyncInterval, grpc.WithInsecure()); err != nil {
			//	log.Printf("error add remote node %s, %s", node, err)
			//}
		}(node)
	}

	go startMux(*addrMain)

	if err := r.StartReplicationServer(ln); err != nil {
		panic(err)
	}
}

func startMux(addrMain string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/upsert", upsert)
	mux.HandleFunc("/all", all)

	if err := http.ListenAndServe(addrMain, mux); err != nil {
		log.Printf("error listen mux, %v", err)
		os.Exit(1)
	}
}

func p(w io.Writer, s string) {
	log.Print(s)
	fmt.Fprint(w, s+"\n")
}

func all(writer http.ResponseWriter, request *http.Request) {
	notExpired, expired := r.All()
	p(writer, "Not Expired")
	for name, value := range notExpired {
		p(writer, fmt.Sprintf("%s : %d", name, value))
	}
	p(writer, "Expired")
	for name, value := range expired {
		p(writer, fmt.Sprintf("%s : %d", name, value))
	}
}

func get(writer http.ResponseWriter, request *http.Request) {
	varName := request.URL.Query().Get("var")
	if varName == "" {
		writer.WriteHeader(400)
		p(writer, "varName is empty")
		return
	}

	value, err := r.Get(varName)

	p(writer, fmt.Sprintf("get %s: %d (%s)", varName, value, err))
}

func upsert(writer http.ResponseWriter, request *http.Request) {
	varName := request.URL.Query().Get("var")
	if varName == "" {
		writer.WriteHeader(400)
		p(writer, "varName is empty")
		return
	}

	value := request.URL.Query().Get("value")
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		writer.WriteHeader(400)
		p(writer, fmt.Sprintf("wrong value, %v", err))
		return

	}

	newValue := r.Upsert(varName, int64(valueInt))
	p(writer, fmt.Sprintf("new value: %d", newValue))
}
