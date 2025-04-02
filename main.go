package main

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jibaru/logv/json"
)

func main() {
	// logsPipe contains the log input from stdin.
	logsPipe := os.Stdin

	// Open the terminal device for interactive input.
	tty, err := getTTY()
	if err != nil {
		// Exit if we cannot open the tty.
		os.Exit(1)
	}
	// Reassign os.Stdin for interactive UI input.
	os.Stdin = tty

	app := tview.NewApplication()
	app.EnableMouse(true)

	list := tview.NewList()
	list.SetSelectedBackgroundColor(tcell.Color52)
	list.ShowSecondaryText(false)
	list.SetBorder(true)
	list.SetTitle("Logs (Select a line to expand, 'q' to quit)")

	// Create a status bar (TextView) to display error messages.
	statusBar := tview.NewTextView()
	statusBar.SetText("")
	statusBar.SetDynamicColors(true)
	statusBar.SetTextAlign(tview.AlignLeft)

	// Global variables for log lines.
	var logLines []string
	var linesMutex sync.Mutex

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

	// Search input field.
	searchInput := tview.NewInputField()
	searchInput.SetLabel("Search: ")
	searchInput.SetFieldWidth(0)
	searchInput.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(list)
	})
	searchInput.SetChangedFunc(func(text string) {
		updateList(text)
	})

	// Clear button to clear the search input.
	clearButton := tview.NewButton("Clear")
	clearButton.SetSelectedFunc(func() {
		searchInput.SetText("")
		updateList("")
	})

	// Top container grouping searchInput and clearButton.
	searchFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	searchFlex.AddItem(searchInput, 3, 0, true)
	searchFlex.AddItem(clearButton, 1, 0, false)

	// Main Flex containing the status bar, search area, and log list.
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(statusBar, 1, 0, false) // status bar on top (1 row)
	flex.AddItem(searchFlex, 4, 0, true)
	flex.AddItem(list, 0, 1, false)

	// When a list item is selected, display an expanded modal with details.
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

		closeButton := tview.NewButton("Close")
		closeButton.SetSelectedFunc(func() {
			app.SetRoot(flex, true).SetFocus(list)
		})

		modalContent := tview.NewFlex().SetDirection(tview.FlexRow)
		modalContent.AddItem(textView, 0, 1, false)
		modalContent.AddItem(closeButton, 3, 0, false)

		// Larger modal: set fixed height (30 rows) for the modal content.
		modal := tview.NewFlex()
		modal.AddItem(nil, 0, 1, false)
		innerFlex := tview.NewFlex().SetDirection(tview.FlexRow)
		innerFlex.AddItem(nil, 0, 1, false)
		innerFlex.AddItem(modalContent, 30, 1, true)
		innerFlex.AddItem(nil, 0, 1, false)
		modal.AddItem(innerFlex, 0, 1, true)
		modal.AddItem(nil, 0, 1, false)

		app.SetRoot(modal, true).SetFocus(closeButton)
	})

	// readPipe reads lines from the provided pipe and adds them to the log list.
	readPipe := func(pipe *os.File) {
		scanner := bufio.NewScanner(pipe)
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
		// Instead of logging errors, update the status bar if an error occurs.
		if err := scanner.Err(); err != nil {
			app.QueueUpdateDraw(func() {
				statusBar.SetText("[red]Error reading logs: " + err.Error() + "[-]")
			})
		}
	}

	// Launch goroutine to read from logsPipe (stdin).
	go readPipe(logsPipe)

	// Capture 'q' or 'Q' key events to quit the application.
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		os.Exit(1)
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
