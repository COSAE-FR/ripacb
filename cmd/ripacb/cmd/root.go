package cmd

import (
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/spf13/cobra"
)

const maskedPassword = "*******"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ripacb",
	Short: "ACB client",
	Long: `Automatic Configuration Backup client

This client can interact with an ACB server.
This client do not use the default pfSense API.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := cliconfig.Config.Validate(); err != nil {
			return err
		}
		if cliconfig.Config.Password == maskedPassword || cliconfig.Config.Password == "" {
			cliconfig.Config.Password = cliconfig.DefaultPassword
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().StringVarP(&cliconfig.Config.ServerURL, "server", "s", cliconfig.DefaultServerURL, "ACB server URL")
	rootCmd.PersistentFlags().StringVarP(&cliconfig.Config.Password, "password", "p", maskedPassword, "ACB decryption password")
	rootCmd.PersistentFlags().StringVarP(&cliconfig.Config.Hostname, "hostname", "n", cliconfig.DefaultHostname, "This firewall hostname")
	rootCmd.PersistentFlags().DurationVarP(&cliconfig.Config.DialTimeout, "check-timeout", "c", cliconfig.DefaultDialTimeout, "HTTP dial timeout used by check requests")
}
