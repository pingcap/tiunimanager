package util

import "fmt"

func AssertNoErr(err error) {
	if err == nil {
	} else {
		panic(err)
	}
}

func Assert(b bool) {
	if b {
	} else {
		panic("assert failed")
	}
}

func AssertWithInfo(b bool, info string) {
	if b {
	} else {
		panic(fmt.Sprintf("assert failed: %s", info))
	}
}
