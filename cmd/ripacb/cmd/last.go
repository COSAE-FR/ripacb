package cmd

import (
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/restore"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var defaultLocation = "/tmp/candidate.xml"

// lastCmd represents the check command
var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Get the last ACB configuration",
	Long: `Get the last ACB configuration.

This client do not use the default pfSense API.
`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := CheckServer()
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		list, error := restore.GetBackups(cliconfig.Config.ServerURL, cliconfig.Config.Hostname, cliconfig.Config.Password)
		if error != nil {
			log.Fatalf("Error: %s", error)
		}
		last := lastRevision(list)
		if last == nil {
			log.Fatalf("No backup found")
		}
		got := make(chan int, 20)
		rev, err := restore.Restore(last.Revision, got)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}
		directory := filepath.Dir(defaultLocation)
		if err := os.MkdirAll(directory, 0755); err != nil {
			log.Fatalf("Error: %s", err)
		}
		if err := os.WriteFile(defaultLocation, rev, 0644); err != nil {
			log.Fatalf("Error: %s", err)
		}
		log.Printf("Configuration written to: %s\n", defaultLocation)
	},
}

func init() {
	rootCmd.AddCommand(lastCmd)
	lastCmd.Flags().DurationVarP(&cliconfig.Config.RequestTimeout, "request-timeout", "r", cliconfig.DefaultRequestTimeout, "HTTP request timeout used by API requests")
	lastCmd.Flags().StringVarP(&defaultLocation, "location", "l", defaultLocation, "Location of the candidate.xml file")
}

func lastRevision(list *entity.RevisionList) *entity.Revision {
	var current *entity.Revision
	for _, rev := range *list {
		if current == nil {
			current = &rev
			continue
		}
		if current.Date.Before(rev.Date) {
			current = &rev
		}
	}
	return current
}
