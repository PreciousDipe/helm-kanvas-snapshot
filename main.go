package main

import (
	cmd "github.com/meshery-extensions/helm-kanvas-snapshot/cmd/kanvas-snapshot"
	"github.com/layer5io/meshkit/logger"
)

var (
	providerToken          string
	mesheryCloudAPIBaseURL string
	mesheryAPIBaseURL      string
	workflowAccessToken    string
	Log                    logger.Handler
)

func main() {
	Log.Infof("email", providerToken)
	cmd.Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken)
}
