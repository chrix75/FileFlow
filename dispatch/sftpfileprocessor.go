package dispatch

import (
	"FileFlow/fileflows"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path"
	"sort"
)

type SFTPFileProcessor struct {
	client *ssh.Client
	sftp   *sftp.Client
}

// Close all resources about SFTP connection
// This method should be defered.
func (p SFTPFileProcessor) Close() {
	p.client.Close()
	p.sftp.Close()
}

// Connect to SFTP server and returns a SFTPFileProcessor for the provided flow.
// flow parameter is the FileFlow description
func Connect(flow fileflows.FileFlow) SFTPFileProcessor {
	client := sshClient(flow, flow.PrivateKeyPath)
	sc := sftpClient(client)

	processor := SFTPFileProcessor{
		client,
		sc,
	}
	return processor
}

// ListFiles list the files in the given directory that match the given pattern of the flow
func (p SFTPFileProcessor) ListFiles(flow fileflows.FileFlow) FileList {
	if _, err := p.sftp.Lstat(flow.SourceFolder); err != nil {
		if os.IsNotExist(err) {
			log.Printf("WARN source folder %s does not exist", flow.SourceFolder)
			return []os.FileInfo{}
		}
	}

	walker := p.sftp.Walk(flow.SourceFolder)
	var files = make(FileList, 0, 50)
	for walker.Step() {
		fileInfo := walker.Stat()
		if !fileInfo.IsDir() && flow.Regexp.MatchString(fileInfo.Name()) {
			files = append(files, fileInfo)
		}
	}
	sort.Sort(files)
	return files
}

// ProcessFile do an action on a file from a SFTP server.
// src parameter is the source full file path in the SFTP server
// dst parameter is the destination full file path in local filesystem
// operation parameter is the operation to do
//
// After the operation is done, the file is moved to the destination folder (so, the file on the SFTP server is removed)
func (p SFTPFileProcessor) ProcessFile(src string, dst string, operation fileflows.FlowOperation) error {

	inp, err := p.sftp.Open(src)
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
		err = copyFile(inp, out)
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

	_ = p.sftp.Remove(src)
	log.Printf("Removed file %s", src)

	return nil
}

// OverflowFile move a file from SFTP to the overflow directory.
// If success, dst contains the full path of the file in the local filesystem.
func (p SFTPFileProcessor) OverflowFile(src string, overflowFolder string) (dst string, err error) {
	inp, err := p.sftp.Open(src)
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

	if _, err := inp.WriteTo(out); err != nil {
		return "", fmt.Errorf("error copying file %s to %s: %v", src, tmp, err)
	}

	dst = ConcatFolderWithFile(overflowFolder, fileName)
	if err := os.Rename(tmp, dst); err != nil {
		return "", fmt.Errorf("error renaming file %s to %s: %v", tmp, dst, err)
	}

	_ = p.sftp.Remove(src)
	return dst, nil
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

func sftpClient(client *ssh.Client) *sftp.Client {
	sc, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal("Failed to sftp: ", err)
	}
	return sc
}

func copyFile(inp *sftp.File, out *os.File) error {
	if _, err := inp.WriteTo(out); err != nil {
		_ = out.Close()
		return err
	}
	return nil
}
