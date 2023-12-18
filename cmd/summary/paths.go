package main

const (
	pathIndex    = "/"
	pathIndexEnv = "/environment/{env}"

	pathAssets = "/assets/"

	pathMetrics = "/metrics"
	pathHealth  = "/health"

	pathUpload   = "/upload"
	pathRadiator = "/radiator"
	pathSearch   = "/search"

	pathNodes    = "/nodes"
	pathNodeFqdn = pathNodes + "/{node_fqdn}"

	pathReports  = "/reports"
	pathReportID = pathReports + "/{report_id}"

	pathApi       = "/api"
	pathApiStates = pathApi + "/states"
	pathApiState  = pathApiStates + "/{state_id}"
)
