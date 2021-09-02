package models

import (
	"testing"
)

func TestDAOManager_InitData(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"normal", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Dao.InitData(); (err != nil) != tt.wantErr {
				t.Errorf("InitData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
