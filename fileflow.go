package main

import (
	"FileFlow/dispatch"
	"FileFlow/fileflows"
	"FileFlow/files"
	"log"
	"sync"
	"time"
)

// FileFlow moves a file from one SFTP directory to another local one.
// To work, a configuration file must be provided that describes all flows.
// The configuration file format is described in the README.md.
func main() {
	var wg sync.WaitGroup

	wg.Add(2)
	flowA := fileflows.NewFileFlow(
		"Move ACME files",
		"localhost",
		22,
		"sftp/acme",
		".+",
		[]string{"/Users/batman/sftp/moved"},
		fileflows.Move,
		3,
		"/Users/batman/sftp/overflow")

	flowB := fileflows.NewLocalFileFlow(
		"Move from ACME overflow folder",
		"/Users/batman/sftp/overflow",
		".+",
		[]string{"/Users/batman/sftp/moved"},
		fileflows.Move,
		3,
		"")

	go func() {
		defer wg.Done()
		for {
			processFlow(flowA, "/Users/batman/.ssh/test.sftp.privatekey.file")
			time.Sleep(time.Second * 10)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			processFlow(flowB, "/Users/batman/.ssh/test.sftp.privatekey.file")
			time.Sleep(time.Second * 10)
		}
	}()

	wg.Wait()
}

func processFlow(flow fileflows.FileFlow, keyFile string) {
	var processor dispatch.FileProcessor
	if flow.IsRemote() {
		remote := dispatch.Connect(flow, keyFile)
		defer remote.Close()
		processor = remote
		log.Printf("Connected to server SFTP for flow %s", flow.Name)
	} else {
		processor = dispatch.Open(flow)
		log.Printf("Start local reading for flow %s", flow.Name)
	}

	allFiles := processor.ListFiles(flow)

	aa := availabilityByFileCount{maxFileCount: flow.MaxFileCount}
	dispatcher := dispatch.NewDispatcher(&flow, dispatch.FolderAvailability(aa), processor)
	for _, f := range allFiles {
		dst, err := dispatcher.Dispatch(f.Name())
		if err != nil {
			log.Printf("WARN cannot move file %s : %v", f.Name(), err)
		} else {
			log.Printf("DEBUG Moved file %s to %s", f.Name(), dst)
		}
	}

}

type availabilityByFileCount struct {
	maxFileCount int
}

func (a availabilityByFileCount) IsAvailable(folder string) bool {
	if a.maxFileCount == 0 {
		return true
	}

	count := files.CountFiles(folder)
	return count > -1 && count < a.maxFileCount
}
