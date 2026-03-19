package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PrintJSON pretty-prints JSON bytes to stdout.
func PrintJSON(data []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}
