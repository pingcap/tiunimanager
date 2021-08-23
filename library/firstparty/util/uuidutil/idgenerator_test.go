package uuidutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateID(t *testing.T) {
	got := GenerateID()
	assert.NotEmpty(t, got)
	assert.True(t, len(got) > 12)
	assert.True(t, len(got) <= 22)
	assert.Less(t, 12, len(got))
	assert.Less(t, len(got), 23)
}
	/*if got == "" {
		t.Errorf("GenerateID() empty, got = %v", got)
	}*/

	/*if len(got) < 12 {
		t.Errorf("GenerateID() too short, got = %v", got)
	}*/

	/*if len(got) > 36 {
		t.Errorf("GenerateID() too long, got = %v", got)
	}
}*/
