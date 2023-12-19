package router

import (
	"net/http"

	authorization "github.com/DX-standarization-Team/common-service-v2/middleware/authorization"
)

const (
	Api1Url      = "https://fika-api1-nakagome-wsgwmfbvhq-uc.a.run.app"
	Api2Url      = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
	workflowUrl  = "https://fika-api2-nakagome-wsgwmfbvhq-uc.a.run.app"
	ProjectId    = "kaigofika-poc01"
	Location     = "us-central1"
	workflowName = "fs-workflow-nakagome"
)

// New sets up our routes and returns a *http.ServeMux
func New() *http.ServeMux {
	router := http.NewServeMux()
	// This route is always accessible.
	router.Handle("/workflow", http.HandlerFunc(workflowHandler))
	// This route is only accessible if the user has a valid access_token.
	router.Handle("/api2", authorization.JwtAuthenticationMiddleware(http.HandlerFunc(api2Handler)))
	return router
}
