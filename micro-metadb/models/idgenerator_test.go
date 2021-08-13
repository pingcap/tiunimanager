package models

import (
	"fmt"
	"testing"
)

func TestGenerateID(t *testing.T) {
	fmt.Println(GenerateID())
	got := GenerateID()
	if got == "" {
		t.Errorf("GenerateID() empty, got = %v", got)
	}

	if len(got) < 12 {
		t.Errorf("GenerateID() too short, got = %v", got)
	}

	if len(got) > 36 {
		t.Errorf("GenerateID() too long, got = %v", got)
	}
}
