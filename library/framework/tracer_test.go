package framework

import (
	"fmt"
	"testing"
)

func TestNewTracerFromArgs(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		args := &ClientArgs{}
		got := NewTracerFromArgs(args)
		fmt.Println(got)
	})
}
