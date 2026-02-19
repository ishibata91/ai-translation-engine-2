package loader

import (
	"encoding/json"
	"fmt"
	"os"
)

// DecodeFile reads the file, detects encoding, and decodes it into a map of RawMessages.
// This is the first phase of the Two-Phase Load strategy.
func DecodeFile(path string) (map[string]json.RawMessage, error) {
	// 1. Open File
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// 2. Detect Encoding and get a UTF-8 converting reader
	reader, err := detectEncoding(f)
	if err != nil {
		return nil, fmt.Errorf("failed to detect encoding: %w", err)
	}

	// 3. Decode into map[string]json.RawMessage (Phase 1)
	// We use DisallowUnknownFields to ensure strict schema compliance if needed,
	// but for now, we just want to load known fields and ignore others if they exist (forward compatibility).
	var rawMap map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&rawMap); err != nil {
		return nil, fmt.Errorf("failed to decode JSON structure: %w", err)
	}

	return rawMap, nil
}
