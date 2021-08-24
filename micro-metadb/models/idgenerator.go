package models

import (
	"encoding/base64"
	"github.com/google/uuid"
)

var ID_LENGTH = 22

func GenerateID() string {
	//return strings.ReplaceAll(uuid.New().String(), "-", "")

	uuid := uuid.New()

	encoded := make([]byte, 22, 22)
	base64.StdEncoding.WithPadding(base64.NoPadding).Encode(encoded, uuid[0:16])

	for i, _ := range encoded {
		if encoded[i] == '/' {
			encoded[i] = '-'
		}
		if encoded[i] == '+' {
			encoded[i] = '*'
		}
	}

	return string(encoded)
}
