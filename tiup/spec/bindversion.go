package spec

// TiEMComponentVersion maps the dm version to the third components binding version
// Empty version means the latest stable one
func TiEMComponentVersion(comp, version string) string {
	switch comp {
	case ComponentAlertmanager,
		ComponentGrafana,
		ComponentPrometheus,
		ComponentBlackboxExporter,
		ComponentNodeExporter,
		ComponentTiEMTracerServer:
		return ""
	case ComponentElasticSearchServer,
		ComponentKibana,
		ComponentTiEMWebServer:
		return ""
	default:
		return version
	}
}
