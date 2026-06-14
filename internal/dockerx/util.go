package dockerx

import (
	"bytes"
	"encoding/json"
)

// indentJSON pretty-prints a raw JSON byte slice, falling back to the original
// text if it cannot be parsed.
func indentJSON(raw []byte) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}
