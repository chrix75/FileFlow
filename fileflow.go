package main

import (
	"FileFlow/dispatch"
	"FileFlow/fileflows"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/fs"
	"log"
	"os"
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
		[]string{"/Users/batman/sftp/moved", "/Users/batman/sftp/moved2"},
		3)
	flowB := fileflows.NewFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		"sftp/nexus",
		".+",
		[]string{"/Users/batman/sftp/moved", "/Users/batman/sftp/moved2"},
		3)

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
	client := sshClient(flow, keyFile)
	defer client.Close()

	sc := sftpClient(client)
	defer sc.Close()

	log.Printf("Connected to server SFTP for flow %s", flow.Name)

	files := files(flow, sc)

	var moveFile dispatch.Callback = func(src string, dst string) error {
		tmpDst := dst + ".tmp"
		log.Printf("INFO Moving %s to %s\n", src, dst)
		inp, err := sc.Open(src)
		if err != nil {
			return err
		}
		defer inp.Close()

		out, err := os.Create(tmpDst)
		if err != nil {
			return err
		}

		if _, err := inp.WriteTo(out); err != nil {
			_ = out.Close()
			return err
		}

		_ = out.Close()
		err = os.Rename(tmpDst, dst)
		if err != nil {
			return err
		}

		_ = sc.Remove(src)

		return nil
	}

	aa := availabilityByFileCount{maxFileCount: flow.MaxFileCount}
	dispatcher := dispatch.NewDispatcher(&flow, dispatch.FolderAvailability(aa), moveFile)
	for _, f := range files {
		dst, err := dispatcher.Dispatch(f.Name())
		if err != nil {
			log.Printf("WARN cannot move file %s to %s: %v", f.Name(), dst, err)
		} else {
			log.Printf("DEBUG Moved file %s to %s", f.Name(), dst)
		}
	}

}

func sftpClient(client *ssh.Client) *sftp.Client {
	sc, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal("Failed to sftp: ", err)
	}
	return sc
}

func files(flow fileflows.FileFlow, sc *sftp.Client) []os.FileInfo {
	walker := sc.Walk(flow.SourceFolder)
	var files = make([]os.FileInfo, 0, 50)
	for walker.Step() {
		fileInfo := walker.Stat()
		if !fileInfo.IsDir() && flow.Regexp.MatchString(fileInfo.Name()) {
			files = append(files, fileInfo)
		}
	}
	return files
}

func sshClient(flow fileflows.FileFlow, keyFile string) *ssh.Client {
	key, err := os.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "batman",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", flow.Server, flow.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	return client
}

type availabilityByFileCount struct {
	maxFileCount int
}

func (a availabilityByFileCount) IsAvailable(folder string) bool {
	if a.maxFileCount == 0 {
		return true
	}

	dir, err := fs.ReadDir(os.DirFS(folder), ".")
	if err != nil {
		return false
	}

	count := 0
	for _, file := range dir {
		if file.Type().IsRegular() {
			count++
		}
	}

	return count < a.maxFileCount
}
