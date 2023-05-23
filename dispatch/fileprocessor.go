package dispatch

import (
	"FileFlow/fileflows"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
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
	ListFiles(flow fileflows.FileFlow) FileList
}

type FileList []os.FileInfo

func (fl FileList) Len() int {
	return len(fl)
}

func (fl FileList) Swap(i, j int) {
	fl[i], fl[j] = fl[j], fl[i]
}

func (fl FileList) Less(i, j int) bool {
	return fl[i].Name() < fl[j].Name()
}

func uncompressOperation(src, dst string, inp fs.File) error {
	tmpDst := dst + ".tmp"
	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}
	defer out.Close()
	if strings.HasSuffix(src, ".gz") {
		finalName := strings.Replace(dst, ".gz", "", 1)
		log.Printf("Decompressing %s to %s", src, finalName)
		err := uncompressFile(inp, out)
		if err != nil {
			return fmt.Errorf("error decompressing file %s to %s: %v", src, tmpDst, err)
		}
		if err := os.Rename(tmpDst, finalName); err != nil {
			return fmt.Errorf("error renaming file %s to %s: %v", tmpDst, finalName, err)
		}
	} else {
		_ = os.Remove(tmpDst)
		return fmt.Errorf("cannot uncompress file %s because it seems to be not compressed", src)
	}

	return nil
}

func compressOperation(src, dst string, inp fs.File) error {
	tmpDst := dst + ".tmp"
	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}
	defer out.Close()
	if !strings.HasSuffix(src, ".gz") {
		gzName := dst + ".gz"
		log.Printf("Compressing %s to %s", src, gzName)
		err := compressFile(inp, out)
		if err != nil {
			return fmt.Errorf("error compressing file %s to %s: %v", src, tmpDst, err)
		}
		if err := os.Rename(tmpDst, gzName); err != nil {
			return fmt.Errorf("error renaming file %s to %s: %v", tmpDst, gzName, err)
		}
	} else {
		_ = os.Remove(tmpDst)
		return fmt.Errorf("cannot compress file %s because it seems to be compressed already", src)
	}

	return nil
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
