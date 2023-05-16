package main

import (
	"FileFlow/dispatch"
	"FileFlow/fileflows"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

// FileFlow moves a file from one SFTP directory to another local one.
// To work, a configuration file must be provided that describes all flows.
// The configuration file format is described in the README.md.
func main() {

	key, err := os.ReadFile("/Users/batman/.ssh/test.sftp.privatekey.file")
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
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

	//todo manage pattern
	pattern := ".+2023"
	flow := fileflows.FileFlow{Name: "Move ACME files", Server: "localhost", Port: 22, SourceFolder: "sftp/acme", Pattern: pattern, DestinationFolders: []string{"/Users/batman/sftp/moved"}}

	addr := fmt.Sprintf("%s:%d", flow.Server, flow.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	sc, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal("Failed to sftp: ", err)
	}
	defer sc.Close()

	fmt.Println("Connected to server SFTP")

	walker := sc.Walk(flow.SourceFolder)
	var files = make([]os.FileInfo, 0, 50)
	for walker.Step() {
		fileInfo := walker.Stat()
		if !fileInfo.IsDir() {
			files = append(files, fileInfo)
		}
	}

	var moveFile dispatch.Callback = func(src string, dst string) error {
		tmpDst := dst + ".tmp"
		fmt.Printf("Moving %s to %s\n", src, dst)
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
			return err
		}

		out.Close()
		err = os.Rename(tmpDst, dst)
		if err != nil {
			return err
		}

		sc.Remove(src)

		return nil
	}
	aa := alwaysAvailable{}
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

type alwaysAvailable struct{}

func (a alwaysAvailable) IsAvailable(_ string) bool {
	return true
}
