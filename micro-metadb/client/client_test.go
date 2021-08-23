package client

import "testing"

func TestInitDBClient(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"normal"},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitDBClient()
		})
	}
}
