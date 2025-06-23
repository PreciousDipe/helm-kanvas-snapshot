package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/layer5io/meshkit/logger"
	"github.com/meshery-extensions/helm-kanvas-snapshot/internal/errors"
	"github.com/meshery-extensions/helm-kanvas-snapshot/internal/log"
	"github.com/spf13/cobra"
)

var (
	ProviderToken          string
	MesheryAPIBaseURL      string
	MesheryCloudAPIBaseURL string
	WorkflowAccessToken    string
	Log                    logger.Handler
)

var (
	chartURI   string
	email      string
	designName string
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

var generateKanvasSnapshotCmd = &cobra.Command{
	Use:   "kanvas",
	Short: "Generate a Kanvas snapshot using a Helm chart",
	Long: `Generate a Kanvas snapshot by providing a Helm chart URI.

		This command allows you to generate a snapshot in Meshery using a Helm chart.

		Example usage:

		helm kanvas-snapshot -f https://meshery.github.io/meshery.io/charts/meshery-v0.7.109.tgz -e your-email@example.com --name nginx-helm

		Flags:
		-f, --file  string	URI to Helm chart (required)
		-e, --email string	email address to notify when snapshot is ready (required)
		    --name  string	(optional) name for the Meshery design
		-h			Help for Helm Kanvas Snapshot plugin`,

	RunE: func(_ *cobra.Command, _ []string) error {
		// Use the extracted name from URI if not provided
		if designName == "" {
			designName = ExtractNameFromURI(chartURI)
			Log.Warnf("No design name provided. Using extracted name: %s", designName)
		}
		if email != "" && !isValidEmail(email) {
			handleError(errors.ErrInvalidEmailFormat(email))
		}
		// Create Meshery Snapshot
		designID, err := CreateMesheryDesign(chartURI, designName, email)
		if err != nil {
			handleError(errors.ErrCreatingMesheryDesign(err))
		}

		assetLocation := fmt.Sprintf("https://raw.githubusercontent.com/layer5labs/meshery-extensions-packages/master/action-assets/helm-plugin-assets/%s.png", designID)

		err = GenerateSnapshot(designID, assetLocation, email, WorkflowAccessToken)
		if err != nil {
			handleError(errors.ErrGeneratingSnapshot(err))
		}

		if email == "" {
			Log.Infof("\nSnapshot generated. Snapshot URL: %s\n", assetLocation)
			Log.Infof("It may take 3-5 minutes for the Kanvas snapshot to display at the above URL.\nTo receive the snapshot via email, use the --email option like this:\n\nhelm helm-kanvas-snapshot -f <chart-URI> [--name <snapshot-name>] [-e <email>]\n")
		} else {
			Log.Infof("\nYou will be notified via email at %s when your Kanvas snapshot is ready.", email)
		}
		return nil
	},
}

type MesheryDesignPayload struct {
	Save  bool   `json:"save"`
	URL   string `json:"url"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ExtractNameFromURI(uri string) string {
	filename := filepath.Base(uri)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func handleError(err error) {
	if err == nil {
		return
	}
	if Log != nil {
		Log.Error(err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(1)
}

func CreateMesheryDesign(uri, name, email string) (string, error) {
	payload := MesheryDesignPayload{
		Save:  true,
		URL:   uri,
		Name:  name,
		Email: email,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		Log.Info("Failed to marshal payload:", err)
		return "", errors.ErrDecodingAPI(err)
	}

	fullURL := fmt.Sprintf("%s/api/pattern/import", MesheryAPIBaseURL)
	// Create the request
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		Log.Info("Failed to create new request:", err)
		return "", errors.ErrHTTPPostRequest(err)
	}

	// Set headers and log them
	req.Header.Set("Cookie", fmt.Sprintf("token=%s;meshery-provider=Layer5", ProviderToken))
	req.Header.Set("Origin", MesheryAPIBaseURL)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.ErrHTTPPostRequest(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.ErrUnexpectedResponseCode(resp.StatusCode, string(body))
	}

	// Decode response
	var result []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.ErrDecodingAPI(fmt.Errorf("failed to decode json. body: %s, error: %w", body, err))
	}

	// Extracting the design ID from the result
	if len(result) > 0 {
		if id, ok := result[0]["id"].(string); ok {
			Log.Infof("Successfully created Meshery design. ID: %s", id)
			return id, nil
		}
	}

	return "", errors.ErrCreatingMesheryDesign(fmt.Errorf("failed to extract design ID from response"))
}

func GenerateSnapshot(contentID, assetLocation, email string, workflowAccessToken string) error {
	payload := fmt.Sprintf(`{"ref":"master","inputs":{"contentID":"%s","assetLocation":"%s", "email":"%s"}}`, contentID, assetLocation, email)
	req, err := http.NewRequest("POST", "https://api.github.com/repos/meshery-extensions/helm-kanvas-snapshot/actions/workflows/kanvas.yaml/dispatches", bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+workflowAccessToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		handleError(errors.ErrGitHubAuth(string(body)))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		handleError(errors.ErrCreatingMesheryDesign(err))
	}

	return nil
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func Main(providerToken, mesheryCloudAPIBaseURL, mesheryAPIBaseURL, workflowAccessToken string) {
	ProviderToken = providerToken
	MesheryCloudAPIBaseURL = mesheryCloudAPIBaseURL
	MesheryAPIBaseURL = mesheryAPIBaseURL
	WorkflowAccessToken = workflowAccessToken
	generateKanvasSnapshotCmd.Flags().StringVarP(&chartURI, "file", "f", "", "URI to Helm chart (required)")
	generateKanvasSnapshotCmd.Flags().StringVar(&designName, "name", "", "Optional name for the Meshery design")
	generateKanvasSnapshotCmd.Flags().StringVarP(&email, "email", "e", "", "Optional email to associate with the Meshery design")

	_ = generateKanvasSnapshotCmd.MarkFlagRequired("file")

	if err := generateKanvasSnapshotCmd.Execute(); err != nil {
		errors.ErrHTTPPostRequest(err)
		generateKanvasSnapshotCmd.SetFlagErrorFunc(func(_ *cobra.Command, _ error) error {
			return nil
		})
	}
}

func init() {
	cobra.OnInitialize(func() {
		Log = log.SetupMeshkitLogger("helm-kanvas-snapshot", false, os.Stdout)
	})
}
