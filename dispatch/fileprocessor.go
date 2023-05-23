package dispatch

import (
	"FileFlow/fileflows"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
)

type FileProcessor interface {
	// ProcessFile do an action on a file
	// Action can be:
	// 1. Copy
	// 2. Compress
	// 3. Decompress
	// src parameter is the source full file path
	// dst parameter is the destination full file path
	// operation parameter is the operation to do
	ProcessFile(src, dst string, operation fileflows.FlowOperation) error

	// OverflowFile move a file to the overflow directory
	// src parameter is the full path of the file to move
	OverflowFile(src, overflowFolder string) (dst string, err error)

	// ListFiles list all the files in the flow's source directory
	ListFiles(flow fileflows.FileFlow) []os.FileInfo
}

func compressFile(inp fs.File, out *os.File) error {
	zw := gzip.NewWriter(out)
	defer zw.Close()

	_, err := io.Copy(zw, inp)
	if err != nil {
		return err
	}

	return nil
}

func uncompressFile(inp fs.File, out *os.File) error {
	r, err := gzip.NewReader(inp)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(out, r)
	if err != nil {
		return err
	}

	return nil
}
