package models

import (
	"github.com/google/uuid"
	"strings"
)

func GenerateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
