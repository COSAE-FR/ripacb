package cmd

import (
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/restore"
	"github.com/spf13/cobra"
)

var ForceConfig bool

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a pfSense configuration.",
	Long: `Restores a previously saved configuration.

This client do not use the default pfSense API.`,
	Run: func(cmd *cobra.Command, args []string) {
		skip := false
		if !ForceConfig {
			_, err := CheckServer()
			skip = err == nil
		}
		rApp := restore.NewApplication(cliconfig.Config, skip)
		if err := rApp.Start(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVarP(&ForceConfig, "force-config", "f", false, "Always show the configuration screen")
	restoreCmd.Flags().DurationVarP(&cliconfig.Config.RequestTimeout, "request-timeout", "r", cliconfig.DefaultRequestTimeout, "HTTP request timeout used by API requests")

}
