package restore

import (
	"fmt"
	"github.com/COSAE-FR/ripacb/cmd/ripacb/cmd/constants"
	"github.com/rivo/tview"
	"strings"
)

const breadcrumbSeparator = " > "

func NewGrid(breadcrumb []string, content tview.Primitive, help string) *tview.Grid {
	breadcrumbText := tview.NewTextView().SetTextAlign(tview.AlignLeft).SetText(fmt.Sprintf(" %s | %s", constants.ApplicationName, strings.Join(breadcrumb, breadcrumbSeparator)))
	helpText := tview.NewTextView().SetTextAlign(tview.AlignLeft).SetText(help)
	grid := tview.NewGrid().
		SetRows(1, 0).
		SetColumns(0, 0).
		SetBorders(true).
		AddItem(breadcrumbText, 0, 0, 1, 2, 0, 0, false).
		AddItem(content, 1, 0, 1, 1, 0, 50, true).
		AddItem(helpText, 1, 1, 1, 1, 0, 0, false)
	return grid
}
