package main

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jibaru/logv/json"
)

func main() {
	// logsPipe contains the log input (typically piped via stdin).
	logsPipe := os.Stdin

	// Open the terminal device for interactive input.
	tty, err := getTTY()
	if err != nil {
		log.Fatalf("Error opening terminal device: %v", err)
	}
	// Reassign os.Stdin for interactive UI input.
	os.Stdin = tty

	app := tview.NewApplication()
	app.EnableMouse(true)

	list := tview.NewList()
	{
		list.SetSelectedBackgroundColor(tcell.Color52)
		list.ShowSecondaryText(false).
			SetBorder(true).
			SetTitle("Logs (Select a line to expand, 'q' to quit)")
	}

	searchInput := tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(list)
		})

	// Arrange the search input and log list vertically.
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(searchInput, 3, 0, true).
		AddItem(list, 0, 1, false)

	var (
		logLines   []string
		linesMutex sync.Mutex
	)

	// updateList refreshes the displayed log list with entries that contain the filter text.
	updateList := func(filter string) {
		linesMutex.Lock()
		defer linesMutex.Unlock()
		list.Clear()
		for _, line := range logLines {
			if strings.Contains(line, filter) {
				var highlightedLine string
				if json.IsValid(line) {
					highlightedLine = json.Highlight(line)
				} else {
					highlightedLine = line
				}
				list.AddItem(highlightedLine, "", 0, nil)
			}
		}
	}

	// Update the log list when the search input changes.
	searchInput.SetChangedFunc(func(text string) {
		updateList(text)
	})

	// Show an expanded view of the selected log line with JSON formatting and highlighting.
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		display := mainText

		textView := tview.NewTextView()
		textView.SetText(display)
		textView.SetBackgroundColor(tcell.Color16)
		textView.SetBorder(true)
		textView.SetTitle("Log Details")
		textView.SetDynamicColors(true)
		textView.SetScrollable(true)
		textView.SetWordWrap(true)

		closeButton := tview.NewButton("Close").SetSelectedFunc(func() {
			app.SetRoot(flex, true).SetFocus(list)
		})

		modalContent := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(textView, 0, 1, false).
			AddItem(closeButton, 1, 0, false)

		modal := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(modalContent, 0, 1, true).
				AddItem(nil, 0, 1, false), 0, 1, true).
			AddItem(nil, 0, 1, false)

		app.SetRoot(modal, true).SetFocus(closeButton)
	})

	// Read log lines asynchronously.
	go func() {
		scanner := bufio.NewScanner(logsPipe)
		for scanner.Scan() {
			line := scanner.Text()
			linesMutex.Lock()
			logLines = append(logLines, line)
			linesMutex.Unlock()
			currentFilter := searchInput.GetText()
			if strings.Contains(line, currentFilter) {
				var highlightedLine string
				if json.IsValid(line) {
					highlightedLine = json.Highlight(line)
				} else {
					highlightedLine = line
				}
				app.QueueUpdateDraw(func() {
					list.AddItem(highlightedLine, "", 0, nil)
				})
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading logs: %v", err)
		}
	}()

	// Capture 'q' or 'Q' key events to quit the application.
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}

// getTTY opens the terminal device for interactive input.
// On Unix systems it uses "/dev/tty", and on Windows it uses "CONIN$".
func getTTY() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return os.Open("CONIN$")
	}
	return os.Open("/dev/tty")
}
