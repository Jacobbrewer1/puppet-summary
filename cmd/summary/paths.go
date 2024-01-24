package main

const (
	pathIndex = "/"
	pathApi   = "/api"

	pathIndexEnv = "/environment/{env}"

	pathAssets = "/assets/"

	pathMetrics = "/metrics"
	pathHealth  = "/health"

	pathUpload = "/upload"

	pathNodes    = "/nodes"
	pathNodeFqdn = pathNodes + "/{node_fqdn}"
	pathNodesEnv = pathNodes + "/environment/{env}"

	pathReports  = "/reports"
	pathReportID = pathReports + "/{report_id}"

	pathStates  = "/states"
	pathStateID = pathStates + "/{state_id}"
)
