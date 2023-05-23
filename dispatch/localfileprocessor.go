package dispatch

import (
	"FileFlow/fileflows"
	"fmt"
	"github.com/kr/fs"
	"io"
	"log"
	"os"
	"path"
)

type LocalFileProcessor struct {
	sourceFolder string
}

// Close is a noop in this context
func (p LocalFileProcessor) Close() {
}

func Open(flow fileflows.FileFlow) LocalFileProcessor {
	return LocalFileProcessor{
		sourceFolder: flow.SourceFolder,
	}
}

// ListFiles list the files in the given directory that match the given pattern of the flow
func (p LocalFileProcessor) ListFiles(flow fileflows.FileFlow) []os.FileInfo {
	if p.sourceFolder != flow.SourceFolder {
		log.Fatal("source folder of the current processor does not match the flow source folder")
	}

	if _, err := os.Stat(p.sourceFolder); err != nil {
		if os.IsNotExist(err) {
			log.Printf("WARN source folder %s does not exist", p.sourceFolder)
			return []os.FileInfo{}
		}
	}

	walker := fs.Walk(p.sourceFolder)
	var files = make([]os.FileInfo, 0, 50)
	for walker.Step() {
		fileInfo := walker.Stat()
		if !fileInfo.IsDir() && flow.Regexp.MatchString(fileInfo.Name()) {
			files = append(files, fileInfo)
		}
	}
	return files
}

// ProcessFile do an action on a file from a local directory
// src parameter is the source full file path
// dst parameter is the destination full file path
// operation parameter is the operation to do
func (p LocalFileProcessor) ProcessFile(src string, dst string, operation fileflows.FlowOperation) error {

	inp, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inp.Close()

	if operation == fileflows.Move {
		tmpDst := dst + ".tmp"
		out, err := os.Create(tmpDst)
		if err != nil {
			return err
		}
		defer out.Close()
		log.Printf("Moving %s to %s", src, dst)
		_, err = io.Copy(out, inp)
		if err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", src, tmpDst, err)
		}
		if err := os.Rename(tmpDst, dst); err != nil {
			return fmt.Errorf("error renaming file %s to %s: %v", tmpDst, dst, err)
		}

	} else if operation == fileflows.Compression {
		if err := compressOperation(src, dst, inp); err != nil {
			return err
		}
	} else if operation == fileflows.Decompression {
		if err := uncompressOperation(src, dst, inp); err != nil {
			return err
		}
	}

	_ = os.Remove(src)
	log.Printf("Removed file %s", src)

	return nil
}

// OverflowFile move a file to the overflow directory.
// If success, dst contains the full path of the file
func (p LocalFileProcessor) OverflowFile(src string, overflowFolder string) (dst string, err error) {
	inp, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %v", src, err)
	}
	defer inp.Close()

	fileName := path.Base(src)
	tmp := ConcatFolderWithFile(overflowFolder, fileName+".tmp")
	out, err := os.Create(tmp)
	if err != nil {
		return "", fmt.Errorf("error creating file %s: %v", tmp, err)
	}

	if _, err := io.Copy(out, inp); err != nil {
		return "", fmt.Errorf("error copying file %s to %s: %v", src, tmp, err)
	}

	dst = ConcatFolderWithFile(overflowFolder, fileName)
	if err := os.Rename(tmp, dst); err != nil {
		return "", fmt.Errorf("error renaming file %s to %s: %v", tmp, dst, err)
	}

	_ = os.Remove(src)
	return dst, nil
}
