package embed

import (
	goembed "embed"
)

//go:embed configs
var embededConfigs goembed.FS

// ReadConfigTemplate read the template file embed.
func ReadConfigTemplate(path string) ([]byte, error) {
	return embededConfigs.ReadFile(path)
}

//go:embed scripts
var embededScripts goembed.FS

// ReadScriptTemplate read the template file embed.
func ReadScriptTemplate(path string) ([]byte, error) {
	return embededScripts.ReadFile(path)
}
