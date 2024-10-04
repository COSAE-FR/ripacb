package restore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/cliconfig"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/constants"
	"github.com/COSAE-FR/ripacb/pkg/acb/bindings"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	"github.com/rivo/tview"
	"io"
	"net/http"
)

type Backups struct {
	App       *Application
	List      *tview.List
	Revisions map[string]entity.Revision
}

func NewBackups(app *Application) *Backups {
	backups := &Backups{App: app, List: tview.NewList()}
	backups.List.SetDoneFunc(func() {
		app.Pages.SwitchToPage(constants.Configuration)
	})
	backups.List.SetSelectedFunc(func(index int, mainLabel string, secondaryLabel string, shortcut rune) {
		progressChan := app.Progress.Init(16, 16)
		app.Pages.SwitchToPage(constants.Progress)
		go func() {
			defer func() {
				close(progressChan)
			}()
			err := restoreBackup(mainLabel, progressChan)
			if err != nil {
				app.Application.QueueUpdateDraw(func() {
					errorModal(app.Pages, constants.Backups, "Error while applying backup:\n%s", err)
				})
				return
			}
			app.Application.QueueUpdateDraw(func() {
				app.Pages.SwitchToPage(constants.Reboot)
			})
		}()
	})
	grid := NewGrid([]string{constants.Configuration, constants.Backups}, backups.List, constants.BackupsHelp)
	app.Pages.AddPage(constants.Backups, grid, true, false)

	return backups
}

func GetBackups(server, hostname, password string) (*entity.RevisionList, error) {
	req := bindings.GetBackupRequest{
		Version:   "22.2",
		DeviceKey: deviceKey(hostname, password),
	}
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: cliconfig.Config.RequestTimeout,
	}
	post, err := client.Post(server+"/api/v1/backups", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if post.Body != nil {
		defer post.Body.Close()
	}
	if post.StatusCode == http.StatusOK {
		raw, err := io.ReadAll(post.Body)
		if err != nil {
			return nil, err
		}
		var revisions entity.RevisionList
		err = json.Unmarshal(raw, &revisions)
		if err != nil {
			return nil, err
		}
		return &revisions, nil

	} else {
		return nil, fmt.Errorf("server returned %s (%d)", post.Status, post.StatusCode)
	}
}

func (b *Backups) Populate() {
	b.List.Clear()
	runes := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'o', 'p', 'q'}
	cnt := 0
	for _, bckp := range b.Revisions {
		b.List.AddItem(bckp.Revision, bckp.Label(), runes[cnt], nil)
		cnt += 1
		if cnt > len(runes) {
			break
		}
	}
}
