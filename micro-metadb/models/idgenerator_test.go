package models

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	got := GenerateID()
	if got == "" {
		t.Errorf("GenerateID() empty, got = %v", got)
	}

	if len(got) != 22 {
		t.Errorf("GenerateID() want len = %v, got = %v", 22, len(got))
	}
}
