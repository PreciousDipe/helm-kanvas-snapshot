package main

import (
	"fmt"
	"os"
	"strings"

	cmd "github.com/meshery-extensions/helm-kanvas-snapshot/cmd/kanvas-snapshot"
)

var (
	providerToken          string
	mesheryCloudAPIBaseURL string
	mesheryAPIBaseURL      string
	workflowAccessToken    string
)

func main() {
	params := map[string]string{
		"providerToken":          providerToken,
		"mesheryCloudAPIBaseURL": mesheryCloudAPIBaseURL,
		"mesheryAPIBaseURL":      mesheryAPIBaseURL,
		"workflowAccessToken":    workflowAccessToken,
	}
	missing := []string{}
	for name, value := range params {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Missing required parameter(s): %s\n", strings.Join(missing, ", "))
		os.Exit(1)
	}
	cmd.Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken)
}
