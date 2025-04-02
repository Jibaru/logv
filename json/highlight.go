package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Highlight applies more comprehensive syntax highlighting to a JSON string.
func Highlight(input string) string {
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(input), &dat); err == nil {
		return highlightMap(dat, 0)
	}

	var listDat []interface{}
	if err := json.Unmarshal([]byte(input), &listDat); err == nil {
		return highlightList(listDat, 0)
	}

	return input
}

func highlightMap(m map[string]interface{}, depth int) string {
	var sb strings.Builder
	sb.WriteString("[white]{[white]")
	first := true
	for key, value := range m {
		if !first {
			sb.WriteString("[white],[white]")
		}
		sb.WriteString("[yellow]\"")
		sb.WriteString(key)
		sb.WriteString("\"[white]: ")
		sb.WriteString(highlightValue(value, depth+1))
		first = false
	}
	sb.WriteString("[white]}[white]")
	return sb.String()
}

func highlightList(list []interface{}, depth int) string {
	var sb strings.Builder
	sb.WriteString("[white][[white]")
	first := true
	for _, value := range list {
		if !first {
			sb.WriteString("[white],[white]")
		}
		sb.WriteString(highlightValue(value, depth+1))
		first = false
	}
	sb.WriteString("[white]][white]")
	return sb.String()
}

func highlightValue(value interface{}, depth int) string {
	switch v := value.(type) {
	case string:
		return "[green]\"" + v + "\"[white]"
	case float64, int, int64:
		return "[blue]" + tview.Escape(strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(fmt.Sprintf("%v", v)), `}`), `]`))) + "[white]"
	case bool:
		return "[cyan]" + fmt.Sprintf("%v", v) + "[white]"
	case nil:
		return "[magenta]null[white]"
	case map[string]interface{}:
		return highlightMap(v, depth)
	case []interface{}:
		return highlightList(v, depth)
	default:
		return tview.Escape(fmt.Sprintf("%v", v))
	}
}
