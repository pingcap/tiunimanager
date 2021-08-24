package uuidutil

import (
	"github.com/lithammer/shortuuid"
)

func GenerateID() string {
	return shortuuid.New()
}
