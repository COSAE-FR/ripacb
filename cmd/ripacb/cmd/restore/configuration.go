package restore

import (
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/constants"
	"github.com/rivo/tview"
)

type Configuration struct {
	App  *Application
	Form *tview.Form
}

func NewConfiguration(app *Application) *Configuration {
	cnf := &Configuration{App: app}
	cnf.Form = tview.NewForm().
		AddInputField(constants.Server, cliconfig.Config.ServerURL, 35, nil, nil).
		AddInputField(constants.Hostname, cliconfig.Config.Hostname, 35, nil, nil).
		AddInputField(constants.Password, cliconfig.Config.Password, 35, nil, nil).
		AddButton(constants.ListBackups, func() {
			cliconfig.Config.ServerURL = cnf.Form.GetFormItemByLabel(constants.Server).(*tview.InputField).GetText()
			cliconfig.Config.Hostname = cnf.Form.GetFormItemByLabel(constants.Hostname).(*tview.InputField).GetText()
			cliconfig.Config.Password = cnf.Form.GetFormItemByLabel(constants.Password).(*tview.InputField).GetText()
			if err := cliconfig.Config.Validate(); err != nil {
				errorModal(app.Pages, "Configuration", "Invalid configuration:\n%s", err)
				return
			}
			if cliconfig.Config.Password == "" || cliconfig.Config.Hostname == "" {
				errorModal(app.Pages, "Configuration", "Invalid configuration:\nPassword and Hostname are mandatory.")
				return
			}
			revisions, err := GetBackups(cliconfig.Config.ServerURL, cliconfig.Config.Hostname, cliconfig.Config.Password)
			if err != nil {
				errorModal(app.Pages, constants.Configuration, "Error when getting backups:\n%s", err)
				return
			}
			app.BackupsPage.Revisions = *revisions
			app.BackupsPage.Populate()
			app.Pages.SwitchToPage(constants.Backups)
		}).
		AddButton(constants.Quit, func() {
			app.Application.Stop()
		})
	grid := NewGrid([]string{constants.Configuration}, cnf.Form, constants.ConfigurationHelp)
	app.Pages.AddPage(constants.Configuration, grid, true, true)
	return cnf
}
