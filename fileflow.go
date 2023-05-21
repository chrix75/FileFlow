package main

import (
	"FileFlow/dispatch"
	"FileFlow/fileflows"
	"compress/gzip"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
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
		fileflows.Compression,
		3)

	flowB := fileflows.NewFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		"sftp/nexus",
		".+",
		[]string{"/Users/batman/sftp/moved", "/Users/batman/sftp/moved2"},
		fileflows.Move,
		3)

	flowC := fileflows.NewFileFlow(
		"Move LexCorp files",
		"localhost",
		22,
		"sftp/lexcorp",
		".+",
		[]string{"/Users/batman/sftp/moved", "/Users/batman/sftp/moved2"},
		fileflows.Decompression,
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

	go func() {
		defer wg.Done()
		for {
			processFlow(flowC, "/Users/batman/.ssh/test.sftp.privatekey.file")
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
		inp, err := sc.Open(src)
		if err != nil {
			return err
		}
		defer inp.Close()

		err = processFile(flow, inp, dst)
		if err != nil {
			return fmt.Errorf("error processing file %s: %v", src, err)
		}

		_ = sc.Remove(src)
		log.Printf("Removed file %s", src)

		return nil
	}

	aa := availabilityByFileCount{maxFileCount: flow.MaxFileCount}
	dispatcher := dispatch.NewDispatcher(&flow, dispatch.FolderAvailability(aa), moveFile)
	for _, f := range files {
		dst, err := dispatcher.Dispatch(f.Name())
		if err != nil {
			log.Printf("WARN cannot move file %s : %v", f.Name(), err)
		} else {
			log.Printf("DEBUG Moved file %s to %s", f.Name(), dst)
		}
	}

}

func processFile(flow fileflows.FileFlow, inp *sftp.File, dst string) error {
	tmpDst := dst + ".tmp"
	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}
	defer out.Close()

	if flow.Operation == fileflows.Move {
		log.Printf("Moving %s to %s", inp.Name(), dst)
		err = copyFile(inp, out)
		if err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", inp.Name(), tmpDst, err)
		}
		if err := os.Rename(tmpDst, dst); err != nil {
			return fmt.Errorf("error renaming file %s to %s: %v", tmpDst, dst, err)
		}

	} else if flow.Operation == fileflows.Compression {
		if !strings.HasSuffix(inp.Name(), ".gz") {
			gzName := dst + ".gz"
			log.Printf("Compressing %s to %s", inp.Name(), gzName)
			err = compressFile(inp, out)
			if err != nil {
				return fmt.Errorf("error compressing file %s to %s: %v", inp.Name(), tmpDst, err)
			}
			if err := os.Rename(tmpDst, gzName); err != nil {
				return fmt.Errorf("error renaming file %s to %s: %v", tmpDst, gzName, err)
			}
		} else {
			_ = os.Remove(tmpDst)
			return fmt.Errorf("cannot compress file %s because it seems to be compressed already", inp.Name())
		}
	} else if flow.Operation == fileflows.Decompression {
		if strings.HasSuffix(inp.Name(), ".gz") {
			finalName := strings.Replace(dst, ".gz", "", 1)
			log.Printf("Decompressing %s to %s", inp.Name(), finalName)
			err = uncompressFile(inp, out)
			if err != nil {
				return fmt.Errorf("error decompressing file %s to %s: %v", inp.Name(), tmpDst, err)
			}
			if err := os.Rename(tmpDst, finalName); err != nil {
				return fmt.Errorf("error renaming file %s to %s: %v", tmpDst, finalName, err)
			}
		} else {
			_ = os.Remove(tmpDst)
			return fmt.Errorf("cannot uncompress file %s because it seems to be not compressed", inp.Name())
		}
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

func copyFile(inp *sftp.File, out *os.File) error {
	if _, err := inp.WriteTo(out); err != nil {
		_ = out.Close()
		return err
	}
	return nil
}

func sftpClient(client *ssh.Client) *sftp.Client {
	sc, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal("Failed to sftp: ", err)
	}
	return sc
}

func files(flow fileflows.FileFlow, sc *sftp.Client) []os.FileInfo {
	if _, err := sc.Lstat(flow.SourceFolder); err != nil {
		if os.IsNotExist(err) {
			log.Printf("WARN source folder %s does not exist", flow.SourceFolder)
			return []os.FileInfo{}
		}
	}

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

	count := countFiles(folder)
	return count > -1 && count < a.maxFileCount
}

func countFiles(folder string) int {
	dir, err := fs.ReadDir(os.DirFS(folder), ".")
	if err != nil {
		return -1
	}

	count := 0
	for _, file := range dir {
		if file.Type().IsRegular() {
			count++
		}
	}

	return count
}
