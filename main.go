package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
				list.AddItem(line, "", 0, nil)
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
		if pretty, err := prettyJSON(mainText); err == nil {
			display = highlightJSON(pretty)
		}
		modal := tview.NewModal().
			SetText(display).
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.SetRoot(flex, true).SetFocus(list)
			})
		app.SetRoot(modal, true).SetFocus(modal)
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
				app.QueueUpdateDraw(func() {
					list.AddItem(line, "", 0, nil)
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

// prettyJSON formats a JSON string with proper indentation.
// It returns an error if the input is not valid JSON.
func prettyJSON(input string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(input), "", "  "); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// highlightJSON applies basic syntax highlighting to a formatted JSON string.
// It highlights keys in yellow and string values in green.
func highlightJSON(input string) string {
	// Highlight keys: e.g. "key": becomes [yellow]"key"[white]:
	keyRegex := regexp.MustCompile(`"([^"]+)":`)
	output := keyRegex.ReplaceAllString(input, `[yellow]"$1"[white]:`)

	// Highlight string values: e.g. : "value" becomes : [green]"value"[white]
	stringValRegex := regexp.MustCompile(`: "([^"]+)"`)
	output = stringValRegex.ReplaceAllString(output, `: [green]"$1"[white]`)

	return output
}
