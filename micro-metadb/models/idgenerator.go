package models

import (
	"github.com/google/uuid"
	"strings"
)

var ID_LENGTH = 32

func GenerateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
	//
	//uuid := uuid.New()
	//
	//encoded := make([]byte, 22, 22)
	//base64.StdEncoding.WithPadding(base64.NoPadding).Encode(encoded, uuid[0:16])
	//
	//return string(encoded)
}
