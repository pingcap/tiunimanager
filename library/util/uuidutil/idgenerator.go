package uuidutil

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid"
)

const (
	ENTITY_UUID_LENGTH int = 22
)

// GenerateID Encode with Base64
// URLEncoding Set contains '-' and '_'
func GenerateID() string {
	uuid := uuid.New()
	encoded := make([]byte, ENTITY_UUID_LENGTH)
	base64.URLEncoding.WithPadding(base64.NoPadding).Encode(encoded, uuid[0:16])

	return string(encoded)
}

// ShortId Encode with Base57
func ShortId() string {
	return shortuuid.New()
}
