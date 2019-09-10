package main

import (
	"time"
)

// call upsert on node1 and read value from node4
// wait 1 sec for sync node1 -> node2 -> node3 -> node4
func case1() error {

	// call Upsert on node1
	body, err := request(1, "upsert?name=VAR-1&value=100")
	if err != nil {
		return err
	}

	if err := assertString(body, "new value: 100", "unexpected body 1"); err != nil {
		return err
	}

	time.Sleep(time.Second)

	// read from node4
	body, err = request(4, "get?name=VAR-1")
	if err != nil {
		return err
	}

	if err := assertString(body, "value: 100", "unexpected body 2"); err != nil {
		return err
	}

	return nil
}
