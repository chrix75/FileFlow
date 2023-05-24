package dispatch

import (
	"FileFlow/fileflows"
	"errors"
	"os"
	"regexp"
	"testing"
)

var noop = noopFileProcessor{}

func TestDispatchOneFileIntoOneDestination(t *testing.T) {
	// Given
	pattern := ".+"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1"}, Regexp: regexp.MustCompile(pattern)}

	var mock FolderAvailability = new(mockAlwaysTrueFolderAvailability)

	// When
	dispatcher := NewDispatcher(&flow, mock, noop)
	dst, err := dispatcher.Dispatch("file_A")

	// Then
	if err != nil {
		t.Errorf("Error dispatching file: %s", err)
	}

	if dst != "/dest1/file_A" {
		t.Errorf("Expected destination: %s, got: %s", "/dest1/file_A", dst)
	}
}

func TestDispatchTwoFileIntoOneDestination(t *testing.T) {
	// Given
	pattern := ".+"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1"}, Regexp: regexp.MustCompile(pattern)}

	var mock FolderAvailability = new(mockAlwaysTrueFolderAvailability)

	// When
	dispatcher := NewDispatcher(&flow, mock, noop)
	dst, err := dispatcher.Dispatch("file_A")
	dst2, err2 := dispatcher.Dispatch("file_B")

	// Then
	if err != nil || err2 != nil {
		t.Errorf("Error dispatching file: %s", err)
	}

	if dst != "/dest1/file_A" {
		t.Errorf("Expected destination: %s, got: %s", "/dest1/file_A", dst)
	}

	if dst2 != "/dest1/file_B" {
		t.Errorf("Expected destination: %s, got: %s", "/dest1/file_B", dst2)
	}
}

func TestDispatchOneFileIntoManyDestinations(t *testing.T) {
	// Given
	pattern := ".+"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1", "/dest2"}, Regexp: regexp.MustCompile(pattern)}

	var mock FolderAvailability = new(mockAlwaysTrueFolderAvailability)

	// When
	dispatcher := NewDispatcher(&flow, mock, noop)
	dst, err := dispatcher.Dispatch("file_A")

	// Then
	if err != nil {
		t.Errorf("Error dispatching file: %s", err)
	}

	if dst != "/dest1/file_A" {
		t.Errorf("Expected destination: %s, got: %s", "/dest1/file_A", dst)
	}
}

func TestDispatchTwoFilesIntoManyDestinations(t *testing.T) {
	// Given
	pattern := ".+"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1", "/dest2"}, Regexp: regexp.MustCompile(pattern)}

	var mock FolderAvailability = new(mockAlwaysTrueFolderAvailability)

	// When
	dispatcher := NewDispatcher(&flow, mock, noop)
	dst, err := dispatcher.Dispatch("file_A")
	dst2, err2 := dispatcher.Dispatch("file_B")

	// Then
	if err != nil || err2 != nil {
		t.Errorf("Error dispatching file: %s", err)
	}

	if dst != "/dest1/file_A" {
		t.Errorf("Expected destination: %s, got: %s", "/dest1/file_A", dst)
	}

	if dst2 != "/dest2/file_B" {
		t.Errorf("Expected destination: %s, got: %s", "/dest2/file_B", dst2)
	}
}

func TestDestinationAvailability(t *testing.T) {
	// Given
	pattern := ".+"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1", "/dest2"}, Regexp: regexp.MustCompile(pattern)}

	// When
	var mock FolderAvailability = new(mockFolderAvailability)
	dispatcher := NewDispatcher(&flow, mock, noop)
	dst, err := dispatcher.Dispatch("file_A")

	// Then
	if err != nil {
		t.Errorf("Error dispatching file: %s", err)
	}

	if dst != "/dest2/file_A" {
		t.Errorf("Expected destination: %s, got: %s", "/dest2/file_A", dst)
	}
}

func TestNoDestinationIsAvailable(t *testing.T) {
	// Given
	pattern := ".+"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1"}, Regexp: regexp.MustCompile(pattern)}

	// When
	var mock FolderAvailability = new(mockFolderAvailability)
	dispatcher := NewDispatcher(&flow, mock, noop)
	_, err := dispatcher.Dispatch("file_A")

	// Then
	if err == nil {
		t.Errorf("Expected error, got: %s", err)
	}

	if !errors.As(err, &DispatcherError{}) {
		t.Errorf("Expected DispatcherError, got: %s", err)
	}
}

type mockAlwaysTrueFolderAvailability struct{}
type mockFolderAvailability struct{}

func (m *mockFolderAvailability) IsAvailable(folder string) bool {
	return folder != "/dest1"
}

func (m *mockAlwaysTrueFolderAvailability) IsAvailable(_ string) bool {
	return true
}

type noopFileProcessor struct{}

func (n noopFileProcessor) ProcessFile(_, _ string, _ fileflows.FlowOperation) error {
	return nil
}

func (n noopFileProcessor) OverflowFile(_, _ string) (dst string, err error) {
	return "/", nil
}

func (n noopFileProcessor) ListFiles(_ fileflows.FileFlow) FileList {
	return []os.FileInfo{}
}
