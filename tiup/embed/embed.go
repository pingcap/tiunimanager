package embed

import (
	goembed "embed"
)

//go:embed templates
var embededTmpls goembed.FS

// ReadTemplates read the template file embed.
func ReadTemplate(path string) ([]byte, error) {
	return embededTmpls.ReadFile(path)
}
