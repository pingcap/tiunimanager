package spec

// TiUniManagerComponentVersion maps the dm version to the third components binding version
// Empty version means the latest stable one
func TiUniManagerComponentVersion(comp, version string) string {
	switch comp {
	case ComponentAlertmanager,
		ComponentGrafana,
		ComponentPrometheus,
		ComponentNodeExporter,
		ComponentFilebeat,
		ComponentTiUniManagerTracerServer:
		return ""
	case ComponentElasticSearchServer,
		ComponentKibana,
		ComponentTiUniManagerWebServer:
		return ""
	default:
		return version
	}
}
