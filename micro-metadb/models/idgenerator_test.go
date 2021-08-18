package models

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	got := GenerateID()
	if got == "" {
		t.Errorf("GenerateID() empty, got = %v", got)
	}

	if len(got) != ID_LENGTH {
		t.Errorf("GenerateID() want len = %v, got = %v", ID_LENGTH, len(got))
	}
}
