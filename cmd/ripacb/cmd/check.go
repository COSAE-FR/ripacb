package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/pkg/acb/bindings"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net"
	"net/http"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check ACB server status",
	Long: `Checks the ACB server status.

This client do not use the default pfSense API.
`,
	Run: func(cmd *cobra.Command, args []string) {
		status, err := CheckServer()
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		log.Printf("ACB server ready: %s", status.Message)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func CheckServer() (*bindings.StatusResponse, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: cliconfig.Config.DialTimeout,
			}).DialContext,
			TLSHandshakeTimeout: cliconfig.Config.DialTimeout * 3,
		},
		Timeout: cliconfig.Config.DialTimeout * 6,
	}
	resp, err := client.Get(cliconfig.Config.ServerURL + "/api/v1/status")
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, err
	}
	status := bindings.StatusResponse{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		return nil, err
	}
	if status.Code == http.StatusOK {
		return &status, nil
	} else {
		return &status, fmt.Errorf("ACB server not ready: %s (%d)", status.Message, status.Code)
	}
}
