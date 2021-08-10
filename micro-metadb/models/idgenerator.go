package models

import (
	"encoding/base64"
	"github.com/google/uuid"
)

func GenerateID() string {
	uuid := uuid.New()

	encoded := make([]byte, 22, 22)
	base64.StdEncoding.WithPadding(base64.NoPadding).Encode(encoded, uuid[0:16])

	return string(encoded)
}
