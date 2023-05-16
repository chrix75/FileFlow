package dispatch

import (
	"FileFlow/fileflows"
	"fmt"
	"strings"
)

type Dispatcher struct {
	flow               *fileflows.FileFlow
	callback           Callback
	dstOffset          int
	folderAvailability folderAvailability
}

type DispatcherError struct {
	source string
}

func (de DispatcherError) Error() string {
	return fmt.Sprintf("can't dispatch file %s because there is no available folder", de.source)
}

type Callback func(src string, dst string) error

type folderAvailability interface {
	isAvailable(folder string) bool
}

func NewDispatcher(flow *fileflows.FileFlow, fa folderAvailability, callback Callback) *Dispatcher {
	return &Dispatcher{
		flow,
		callback,
		0,
		fa,
	}
}

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
		if d.dstOffset >= len(d.flow.DestinationFolder) {
			d.dstOffset = 0
		}

		if d.dstOffset == start {
			return "", DispatcherError{fileName}
		}
	}
}

func ConcatFolderWithFile(folder string, fileName string) string {
	if strings.HasSuffix(folder, "/") {
		return folder + fileName
	}
	return folder + "/" + fileName
}

func (d *Dispatcher) tryDispatch(fileName string) (string, error) {
	src := ConcatFolderWithFile(d.flow.SourceFolder, fileName)

	folder := d.flow.DestinationFolder[d.dstOffset]
	if d.folderAvailability.isAvailable(folder) {
		dst := ConcatFolderWithFile(folder, fileName)
		if err := d.callback(src, dst); err != nil {
			return "", err
		}

		d.dstOffset++
		if d.dstOffset >= len(d.flow.DestinationFolder) {
			d.dstOffset = 0
		}

		return dst, nil
	}

	return "", nil
}
