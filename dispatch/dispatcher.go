// Package dispatch provide functions and types for dispatching files into different folders.
package dispatch

import (
	"FileFlow/fileflows"
	"FileFlow/files"
	"fmt"
	"strings"
)

// Dispatcher contains data and functions for dispatching files into different folders.
type Dispatcher struct {
	flow *fileflows.FileFlow
	FileProcessor
	dstOffset          int
	folderAvailability FolderAvailability
}

// DispatcherError is an error type for managing error while dispatching files.
type DispatcherError struct {
	source string
}

func (de DispatcherError) Error() string {
	return fmt.Sprintf("can't dispatch file %s because there is no available folder", de.source)
}

// Callback is a function that is called when a destination folder is allocated to a file while dispatching.
// The src parameter is the absolute source file path name and the dst parameter is the absolute destination file path.
type Callback func(src string, dst string) error

// FolderAvailability is a type setting the availability of a folder.
type FolderAvailability interface {
	IsAvailable(folder string) bool
}

// NewDispatcher creates a new Dispatcher instance.
// Example:
//
//	pattern := ".+"
//	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/dest1"}, Regexp: regexp.MustCompile(pattern)}
//
//	callback := func(source string, destination string) error {
//		return nil
//	}
//
//	var mock FolderAvailability = new(mockAlwaysTrueFolderAvailability)
//
//	dispatcher := NewDispatcher(&flow, mock, callback)
func NewDispatcher(flow *fileflows.FileFlow, fa FolderAvailability, processor FileProcessor) *Dispatcher {
	return &Dispatcher{
		flow,
		processor,
		0,
		fa,
	}
}

// Dispatch method dispatches a file into a destination folder.
// This method searches a available folder (using FolderAvailability interface) for the fileName file.
// The src parameter is not an absolute file path but really the file name. The source folder is set in the flow field
// of the Dispatcher instance. When a folder is found, then the callback function is called.
// If the dispatch is successful, then the dst parameter is set to the absolute destination file path and err is nil.
// If any error occurs, then the dst parameter is set to an empty string and err is set.
func (d *Dispatcher) Dispatch(fileName string) (dst string, err error) {
	start := d.dstOffset

	for {
		dst, err := d.tryDispatch(fileName)
		if err != nil {
			return "", err
		}

		if dst != "" {
			return dst, nil
		}

		d.dstOffset++
		if d.dstOffset >= len(d.flow.DestinationFolders) {
			d.dstOffset = 0
		}

		if d.dstOffset == start {
			return "", DispatcherError{fileName}
		}
	}
}

// ConcatFolderWithFile is an utility function that concatenates a folder and a file name.
// It works only for Linux style file path.
func ConcatFolderWithFile(folder string, fileName string) string {
	if strings.HasSuffix(folder, "/") {
		return folder + fileName
	}
	return folder + "/" + fileName
}

func (d *Dispatcher) tryDispatch(fileName string) (string, error) {
	src := ConcatFolderWithFile(d.flow.SourceFolder, fileName)

	folder := d.flow.DestinationFolders[d.dstOffset]
	if overflowFolderIsEmpty(d.flow.OverflowFolder) && d.folderAvailability.IsAvailable(folder) {
		dst := ConcatFolderWithFile(folder, fileName)
		if err := d.ProcessFile(src, dst, d.flow.Operation); err != nil {
			return "", err
		}

		d.dstOffset++
		if d.dstOffset >= len(d.flow.DestinationFolders) {
			d.dstOffset = 0
		}

		return dst, nil
	}

	if d.flow.OverflowFolder != "" {
		dst, err := d.OverflowFile(src, d.flow.OverflowFolder)
		if err != nil {
			return "", fmt.Errorf("move to overflow folder: %w failed", err)
		}

		return dst, nil
	}

	return "", nil
}

func overflowFolderIsEmpty(folder string) bool {
	if folder == "" {
		return true
	}
	return files.CountFiles(folder) == 0
}
