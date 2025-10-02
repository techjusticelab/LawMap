package config

import (
    "encoding/json"
    "fmt"
    "os"
)

// SourceDescriptor corresponds to entries in configs/sources.json.
type SourceDescriptor struct {
    Name          string   `json:"name"`
    Jurisdictions []string `json:"jurisdictions"`
    Codes         []string `json:"codes"`
    Kind          string   `json:"kind"`   // bulk|api|web|mixed
    URLs          []string `json:"urls"`
}

// LoadSources reads a JSON array of SourceDescriptor from the provided path.
func LoadSources(path string) ([]SourceDescriptor, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read sources config: %w", err)
    }
    var out []SourceDescriptor
    if err := json.Unmarshal(b, &out); err != nil {
        return nil, fmt.Errorf("parse sources config: %w", err)
    }
    return out, nil
}

