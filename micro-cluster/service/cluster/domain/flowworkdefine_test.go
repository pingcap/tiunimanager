package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_defaultContextParser(t *testing.T) {
	ctx := defaultContextParser("")
	assert.NotNil(t, ctx)
}