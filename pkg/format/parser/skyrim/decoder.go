package skyrim

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// DecodeFile reads the file, detects encoding, and decodes it into a map of RawMessages.
// This is the first phase of the Two-Phase Load strategy.
func DecodeFile(path string) (map[string]json.RawMessage, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, fmt.Errorf("open parser input path=%s: %w", path, err)
	}
	defer f.Close()

	reader, err := createUTF8Reader(f)
	if err != nil {
		return nil, fmt.Errorf("create utf8 reader path=%s: %w", path, err)
	}

	return decodeJSON(reader)
}

// openFile opens the specified file for reading.
func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}

// createUTF8Reader detects encoding and returns a UTF-8 converting reader.
func createUTF8Reader(r io.ReadSeeker) (io.Reader, error) {
	reader, err := detectEncoding(r)
	if err != nil {
		return nil, fmt.Errorf("failed to detect encoding: %w", err)
	}
	return reader, nil
}

// decodeJSON decodes the reader content into a map of json.RawMessage.
func decodeJSON(reader io.Reader) (map[string]json.RawMessage, error) {
	var rawMap map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&rawMap); err != nil {
		return nil, fmt.Errorf("failed to decode JSON structure: %w", err)
	}

	return rawMap, nil
}
