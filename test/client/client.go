package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type caseFunc func() error

var (
	cases = map[string]caseFunc{
		"upsert on node1 and read from node4": case1,
		"upsert on different nodes":           case2,
	}
)

func main() {
	for n, c := range cases {

		fmt.Printf("start case '%s' ", n)

		if err := c(); err != nil {
			fmt.Printf("FAIL\n")
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("OK\n")
	}

	fmt.Printf("all ok\n")
}

func request(nodeID int, query string) (string, error) {

	url := "http://127.0.0.1:200" + strconv.Itoa(nodeID) + "/" + query

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error send request, %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error read body, %v", err)
	}

	return string(body), nil
}
