package restore

import (
	"fmt"
	"github.com/rivo/tview"
	"strings"
)

type Progress struct {
	textView *tview.TextView
	app      *tview.Application
	Label    string
	full     int
	limit    int
	progress chan int
}

// Init initialize the progress bar
// full is the maximum amount of value can be sent to channel
// limit is the progress bar size
func (p *Progress) Init(full int, limit int) chan int {
	p.progress = make(chan int, full)
	p.full = full
	p.limit = limit

	go func() { // Simple channel status gauge (progress bar)
		progress := 0
		for inc := range p.progress {
			progress += inc

			if progress > full {
				break
			}

			x := progress * limit / full
			p.app.QueueUpdate(func() {
				p.textView.Clear()
			})
			_, _ = fmt.Fprintf(p.textView, "%s:  %s%s %d/%d",
				p.Label,
				strings.Repeat("■", x),
				strings.Repeat("□", limit-x),
				progress, full)
		}
	}()
	return p.progress
}
