package main

import (
	"fmt"
	"time"
)

// call upsert on node1, call upsert on node2 and get updated from node1
func case2() error {

	// call Upsert on node1
	body, err := request(1, "upsert?name=VAR-2&value=100")
	if err != nil {
		return err
	}

	if err := assertString(body, "new value: 100", "unexpected body 1"); err != nil {
		return err
	}

	time.Sleep(time.Second)

	// call Upsert on node2
	body, err = request(2, "upsert?name=VAR-2&value=-150")
	if err != nil {
		return err
	}

	if err := assertString(body, "new value: -50", "unexpected body 2"); err != nil {
		return err
	}
	if body != "new value: -50" {
		return fmt.Errorf("unexpected body 2 '%s", body)
	}

	time.Sleep(time.Second)

	// read from node1
	body, err = request(1, "get?name=VAR-2")
	if err != nil {
		return err
	}

	if err := assertString(body, "value: -50", "unexpected body 3"); err != nil {
		return err
	}

	return nil
}
