//nolint:revive // utils is an acceptable package name for utility functions
package utils

import (
	"encoding/json"
	"fmt"
)

// ParseJSON parses JSON bytes into the target interface
func ParseJSON(data []byte, target interface{}) error {
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}
