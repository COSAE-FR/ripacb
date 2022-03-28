package constants

const ApplicationName = "ACB Restore"

const (
	Configuration = "Configuration"
	Quit          = "Quit"
	Server        = "Server"
	Hostname      = "Hostname"
	Password      = "Password"
	Backups       = "Backups"
	ListBackups   = "List backups"
	Reboot        = "Reboot"
	Progress      = "Progress"
	Restore       = "Restore"
)

const ConfigurationHelp = `
The Server field must be the ACB server URL, https://acb.example.com.

The Hostname is your firewall hostname.

The Password field is your backup decryption password.

Configure your backup server here.
`
const BackupsHelp = `
Choose the backup to restore and press Enter to proceed.

Press Escape to go to the Configuration page.
`
