package knowledge

import (
	"testing"
)

func TestMain(m *testing.M) {
	LoadKnowledge()
	m.Run()
}