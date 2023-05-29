package main

import (
	"FileFlow/dispatch"
	"FileFlow/fileflows"
	"FileFlow/files"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

// FileFlow moves a file from one SFTP directory to another local one.
// To work, a configuration file must be provided that describes all flows.
// The configuration file format is described in the README.md.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: FileFlow <config file>")
		os.Exit(1)
	}

	configFile := os.Args[1]

	config, err := fileflows.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	processing := true

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				log.Printf("FileFlow is shutting down...")
				processing = false
			}
		}
	}()

	for _, flow := range config.FileFlows {
		wg.Add(1)

		currentFlow := flow
		go func() {
			defer wg.Done()
			for {
				if !processing {
					break
				}
				processFlow(currentFlow)
				time.Sleep(time.Duration(int(time.Second) * config.Delay))
			}
			log.Printf("Flow %s finished", currentFlow.Name)
		}()

	}

	wg.Wait()
	log.Printf("All flows finished.")
}

func processFlow(flow fileflows.FileFlow) {
	var processor dispatch.FileProcessor
	if flow.IsRemote() {
		remote := dispatch.Connect(flow)
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
