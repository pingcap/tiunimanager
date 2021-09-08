package spec

import "github.com/pingcap/tiup/pkg/cluster/spec"

// TiEMComponentVersion maps the dm version to the third components binding version
// Empty version means the latest stable one
func TiEMComponentVersion(comp, version string) string {
	switch comp {
	case spec.ComponentAlertmanager,
		spec.ComponentGrafana,
		spec.ComponentPrometheus,
		spec.ComponentBlackboxExporter,
		spec.ComponentNodeExporter:
		return ""
	case ComponentElasticSearchServer:
		return ""
	default:
		return version
	}
}
