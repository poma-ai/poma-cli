package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// requireToken returns an error if the token is empty.
func requireToken(token string) error {
	if token == "" {
		return fmt.Errorf("token is required (--token or POMA_API_KEY)")
	}
	return nil
}

// PrintJSON pretty-prints JSON bytes to stdout.
func PrintJSON(data []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}
