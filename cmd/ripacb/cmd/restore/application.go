package restore

import (
	"fmt"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/constants"
	"github.com/rivo/tview"
)

type Application struct {
	Application   *tview.Application
	Pages         *tview.Pages
	Configuration cliconfig.Configuration
	ConfPage      *Configuration
	BackupsPage   *Backups
	Progress      *Progress
}

func NewApplication(configuration cliconfig.Configuration, skipConfig bool) *Application {
	app := &Application{
		Configuration: configuration,
		Application:   tview.NewApplication(),
		Pages:         tview.NewPages(),
	}
	app.Pages.SetChangedFunc(func() {
		_, primitive := app.Pages.GetFrontPage()
		if !primitive.HasFocus() {
			app.Application.SetFocus(primitive)
		}
	})
	app.ConfPage = NewConfiguration(app)
	app.BackupsPage = NewBackups(app)

	if skipConfig && cliconfig.Config.ServerURL != "" && cliconfig.Config.Hostname != "" && cliconfig.Config.Password != "" {
		backups, err := GetBackups(cliconfig.Config.ServerURL, cliconfig.Config.Hostname, cliconfig.Config.Password)
		if err == nil {
			app.BackupsPage.Revisions = *backups
			app.BackupsPage.Populate()
			app.Pages.SwitchToPage(constants.Backups)
		}
	}

	// Reboot question modal
	reboot := tview.NewModal()
	reboot.AddButtons([]string{constants.Reboot, constants.Quit})
	reboot.SetText("Restoration finished!\nDo you want to reboot the firewall now?")
	reboot.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonLabel {
		case constants.Reboot:
			fmt.Println("Rebooting, please wait...")
			go func() {
				Reboot()
			}()
		}
		app.Application.Stop()
	})
	app.Pages.AddPage(constants.Reboot, reboot, false, false)

	// Restoration progress modal
	textView := tview.NewTextView().
		SetChangedFunc(func() {
			app.Application.Draw()
		})
	textView.SetBorder(false)
	app.Progress = &Progress{textView: textView, Label: constants.Restore, app: app.Application}
	app.Pages.AddPage(constants.Progress, textView, true, false)
	return app
}

func (a *Application) Start() error {
	return a.Application.SetRoot(a.Pages, true).SetFocus(a.Pages).Run()
}
