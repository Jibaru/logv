package json

import "encoding/json"

// IsValid checks if a string is a valid JSON.
func IsValid(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
