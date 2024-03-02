package web

const (
	pathIndex    = "/"
	pathIndexEnv = "/environment/{env}"

	pathNodes    = "/nodes"
	pathNodeFqdn = pathNodes + "/{node_fqdn}"

	pathReports  = "/reports"
	pathReportID = pathReports + "/{report_id}"
)
