package common_test

import (
	"blink-liveview-websocket/common"
	"fmt"
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGetFingerprintExisting(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, common.FINGERPRINT_FILE)
	os.WriteFile(mockFile, []byte("this-is-a-fake-uuid"), 0644)

	fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, fingerprint.New, false)
	assert.Equal(t, fingerprint.Value, "this-is-a-fake-uuid")
	assert.Equal(t, err, nil)
}

func TestGetFingerprintNew(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, common.FINGERPRINT_FILE)
	fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, fingerprint.New, true)
	assert.NotEqual(t, fingerprint.Value, "") // should be a UUID
	assert.Equal(t, err, nil)
}

func TestGetFingerprintEmptyFile(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, common.FINGERPRINT_FILE)
	os.WriteFile(mockFile, []byte(""), 0644)

	fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, fingerprint.New, true)
	assert.NotEqual(t, fingerprint.Value, "") // should be a UUID
	assert.Equal(t, err, nil)
}

func TestGetFingerprintDefaultFilename(t *testing.T) {
	fingerprint, err := common.GetFingerprint("")

	assert.Equal(t, fingerprint.Filename, common.FINGERPRINT_FILE)
	assert.Equal(t, err, nil)
}

func TestGetFingerprintCustomFilename(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, "custom-fingerprint.txt")

	os.WriteFile(mockFile, []byte("custom file"), 0644)

	fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, fingerprint.Value, "custom file")
	assert.Equal(t, fingerprint.Filename, mockFile)
	assert.Equal(t, err, nil)
}

func TestDestroyFingerprint(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, "test-delete-fingerprint.txt")
	os.WriteFile(mockFile, []byte("custom file"), 0644)

	fingerprint := &common.Fingerprint{
		New:      false,
		Value:    "custom file",
		Filename: mockFile,
	}

	_, err := os.Stat(mockFile)
	assert.Equal(t, err, nil) // File exists

	err = fingerprint.Destroy()
	assert.Equal(t, err, nil) // No error

	_, err = os.Stat(mockFile)
	assert.NotEqual(t, err, nil) // File does not exist
}

func TestDestroyFingerprintNew(t *testing.T) {
	fingerprint := &common.Fingerprint{
		New:      true,
		Value:    "custom file",
		Filename: "mock-file.txt",
	}

	err := fingerprint.Destroy()
	assert.Equal(t, err, nil) // No error when destroying a new fingerprint
}

func TestDestroyFingerprintEmptyFilename(t *testing.T) {
	fingerprint := &common.Fingerprint{
		New:      false,
		Value:    "",
		Filename: "",
	}

	err := fingerprint.Destroy()
	assert.Equal(t, err, fmt.Errorf("fingerprint filename is empty"))
}

func TestStoreFingerprint(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, "test-store-fingerprint.txt")

	fingerprint := &common.Fingerprint{
		New:      true, // New fingerprint
		Value:    "custom file",
		Filename: mockFile,
	}

	_, err := os.Stat(mockFile)
	assert.NotEqual(t, err, nil) // File does not exist

	err = fingerprint.Store()
	assert.Equal(t, err, nil) // No error

	file, err := os.ReadFile(mockFile)
	assert.Equal(t, string(file), "custom file")
	assert.Equal(t, err, nil)
}

func TestStoreFingerprintExisting(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := fmt.Sprintf("%s/%s", mockdir, "test-store-fingerprint.txt")
	os.WriteFile(mockFile, []byte("existing file"), 0644)

	fingerprint := &common.Fingerprint{
		New:      false,      // Existing fingerprint
		Value:    "new file", // Should not be written
		Filename: mockFile,
	}

	err := fingerprint.Store()
	assert.Equal(t, err, nil) // No error

	file, err := os.ReadFile(mockFile)
	assert.Equal(t, string(file), "existing file")
	assert.Equal(t, err, nil)
}

func TestStoreFingerprintEmptyFilename(t *testing.T) {
	fingerprint := &common.Fingerprint{
		New:      true,
		Value:    "custom file",
		Filename: "",
	}

	err := fingerprint.Store()
	assert.Equal(t, err, fmt.Errorf("fingerprint filename is empty"))
}

func TestString(t *testing.T) {
	fingerprint := &common.Fingerprint{
		New:      true,
		Value:    "custom file",
		Filename: "mock-file.txt",
	}

	assert.Equal(t, fingerprint.String(), "custom file")
}

func TestStringEmpty(t *testing.T) {
	fingerprint := &common.Fingerprint{
		New:      true,
		Value:    "",
		Filename: "mock-file.txt",
	}

	assert.Equal(t, fingerprint.String(), "")
}
