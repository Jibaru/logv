package json

import (
	"bytes"
	"encoding/json"
)

// Pretty formats a JSON string with proper indentation.
// It returns an error if the input is not valid JSON.
func Pretty(input string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(input), "", "  "); err != nil {
		return "", err
	}
	return buf.String(), nil
}
