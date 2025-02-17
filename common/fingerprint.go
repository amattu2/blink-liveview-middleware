package common

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type Fingerprint struct {
	// A boolean flag indicating if this is a new fingerprint
	New bool `json:"new"`
	// The fingerprint value
	Value string `json:"value"`
	// The filename of the fingerprint file
	Filename string `json:"filename"`
}

var FINGERPRINT_FILE string = "fingerprint.txt"

// GetFingerprint returns the fingerprint from the fingerprint file.
// If the file does not exist, it will be created and a new fingerprint will be generated.
//
// filename: the filename to use for the fingerprint file. Optional.
//
// Example: GetFingerprint() = &Fingerprint{New: true, Value: "fingerprint"}, nil
func GetFingerprint(filename string) (*Fingerprint, error) {
	if filename == "" {
		filename = FINGERPRINT_FILE
	}

	file, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		return generateNew(filename), nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading fingerprint file: %w", err)
	}

	fingerprint := string(file)
	if fingerprint == "" {
		return generateNew(filename), nil
	}

	return &Fingerprint{
		New:      false,
		Value:    fingerprint,
		Filename: filename,
	}, nil
}

// Destroy removes the fingerprint file from the filesystem
// If the fingerprint is new, nil is returned
// This is not reversible, and should only be used when the fingerprint is no longer needed
//
// Example: Destroy() = nil
func (f *Fingerprint) Destroy() error {
	if f.New {
		return nil
	}
	if f.Filename == "" {
		return fmt.Errorf("fingerprint filename is empty")
	}

	return os.Remove(f.Filename)
}

// Store writes the fingerprint to the filesystem if it is new
//
// Example: Store() = nil
func (f *Fingerprint) Store() error {
	if !f.New {
		return nil
	}
	if f.Filename == "" {
		return fmt.Errorf("fingerprint filename is empty")
	}

	return os.WriteFile(f.Filename, []byte(f.Value), 0644)
}

// String returns the fingerprint value as a string
//
// Example: String() = "fingerprint-xyz"
func (f *Fingerprint) String() string {
	return f.Value
}

// generateNew generates a new fingerprint and returns it.
// It does not write the fingerprint to the filesystem.
//
// Example: generateNew() = &Fingerprint{New: true, Value: "fingerprint", Filename: ""}
func generateNew(filename string) *Fingerprint {
	return &Fingerprint{
		New:      true,
		Value:    uuid.New().String(),
		Filename: filename,
	}
}
