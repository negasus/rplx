package main

import "fmt"

func assertString(real, expect, mes string) error {

	if real != expect {
		return fmt.Errorf("%s: Real: '%s', Expect: '%s'", mes, real, expect)
	}

	return nil
}
