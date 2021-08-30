package models

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	got := GenerateID()
	if got == "" {
		t.Errorf("GenerateID() empty, got = %v", got)
	}

	if len(got) != UUID_MAX_LENGTH {
		t.Errorf("GenerateID() want len = %d, got = %v", UUID_MAX_LENGTH, len(got))
	}

}

func TestGenerateIDReplace(t *testing.T) {
	time := 0
	for time < 100 {
		got := GenerateID()
		if strings.Contains(got, "/") {
			t.Errorf("GenerateID() got /")
		}
		if strings.Contains(got, "-") {
			break
		}
		time++
	}
	for time < 200 {
		got := GenerateID()
		if strings.Contains(got, "/") {
			t.Errorf("GenerateID() got /")
		}
		if strings.Contains(got, "-") {
			break
		}
		time++
	}

}
