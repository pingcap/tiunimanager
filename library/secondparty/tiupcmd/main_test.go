package main

import "testing"

func Test_Main(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	main()
}

