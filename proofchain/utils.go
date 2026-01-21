package proofchain

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// readFile reads a file and returns its contents.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// filepath returns the base name of a path.
func filepathBase(path string) string {
	return filepath.Base(path)
}

// jsonMarshal marshals a value to JSON.
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
