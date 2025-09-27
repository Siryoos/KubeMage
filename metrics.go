// metrics.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metrics struct {
	TotalCommands     int `json:"total_commands"`
	ValidatedCommands int `json:"validated_commands"`
	ExecutedCommands  int `json:"executed_commands"`
}

func (m *Metrics) Dump() {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling metrics: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
