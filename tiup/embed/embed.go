package embed

import (
	goembed "embed"
)

//go:embed templates
var embededTmpls goembed.FS

// ReadTemplate read the template file embed.
func ReadTemplate(path string) ([]byte, error) {
	return embededTmpls.ReadFile(path)
}
