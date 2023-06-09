package main

import (
	"FileFlow/fileflows"
	"FileFlow/files"
	"compress/gzip"
	"io"
	"log"
	"os"
	"testing"
)

var (
	localSftpFolder     = "/Users/batman/sftp/tests/input/"
	localDestFolder     = "/Users/batman/sftp/tests/output/"
	localOverflowFolder = "/Users/batman/sftp/tests/overflow/"
	sftpPrivateKeyFile  = "/Users/batman/.ssh/test.sftp.privatekey.file"

	remoteInputSftpFolder = "sftp/tests/input/"
)

// Integration test
func TestMoveFileFromSftp(t *testing.T) {
	// Given
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localSftpFolder, "file.txt")
	expectedResultFile := localDestFolder + "file.txt"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(expectedResultFile)
	}()

	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Move,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(expectedResultFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("File should be found: %s", expectedResultFile)
		}
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be found: %s", sourceFile)
	}
}

// Integration test
func TestCompressFileFromSftp(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localSftpFolder, "file.txt")
	expectedResultFile := localDestFolder + "file.txt.gz"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(expectedResultFile)
	}()

	// Given
	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Compression,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(expectedResultFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("File should be found: %s", expectedResultFile)
		}
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be found: %s", sourceFile)
	}
}

// Integration test
func TestCompressFileFromLocal(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localSftpFolder, "file.txt")
	expectedResultFile := localDestFolder + "file.txt.gz"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(expectedResultFile)
	}()

	// Given
	flow := fileflows.NewLocalFileFlow(
		"Move Nexus files",
		localSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Compression,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(expectedResultFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("File should be found: %s", expectedResultFile)
		}
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be found: %s", sourceFile)
	}
}

// Integration test
func TestCancelCompressFileFromSftp(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createGzipFile(localSftpFolder, "file.txt")
	unexpectedResultFile := localDestFolder + "file.txt.gz.gz"
	unexpectedTmpFile := localDestFolder + "file.txt.gz.tmp"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(unexpectedResultFile)
		_ = os.Remove(unexpectedTmpFile)
	}()

	// Given
	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Compression,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(unexpectedResultFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedResultFile)
	}

	if _, err := os.Stat(unexpectedTmpFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedTmpFile)
	}

	if _, err := os.Stat(sourceFile); err != nil {
		t.Errorf("File should be found: %s", sourceFile)
	}
}

// Integration test
func TestUncompressFileFromSftp(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createGzipFile(localSftpFolder, "file.txt")
	expectedResultFile := localDestFolder + "file.txt"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(expectedResultFile)
	}()

	// Given
	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Decompression,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(expectedResultFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("File should be found: %s", expectedResultFile)
		}
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be found: %s", sourceFile)
	}
}

// Integration test
func TestUncompressFileFromLocal(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createGzipFile(localSftpFolder, "file.txt")
	expectedResultFile := localDestFolder + "file.txt"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(expectedResultFile)
	}()

	// Given
	flow := fileflows.NewLocalFileFlow(
		"Move Nexus files",
		localSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Decompression,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(expectedResultFile); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("File should be found: %s", expectedResultFile)
		}
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be found: %s", sourceFile)
	}
}

// Integration test
func TestCancelUncompressFileFromSftp(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localSftpFolder, "file.txt")
	unexpectedResultFile := localDestFolder + "file.txt"
	unexpectedTmpFile := localDestFolder + "file.txt.tmp"
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(unexpectedResultFile)
		_ = os.Remove(unexpectedTmpFile)
	}()

	// Given
	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Decompression,
		3,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(unexpectedResultFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedResultFile)
	}

	if _, err := os.Stat(unexpectedTmpFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedTmpFile)
	}

	if _, err := os.Stat(sourceFile); err != nil {
		t.Errorf("File should be found: %s", sourceFile)
	}
}

// Integration test
func TestMoveToOverflowDir(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localSftpFolder, "file.txt")
	unexpectedResultFile := localDestFolder + "file.txt"
	expectedOverflowFile := localOverflowFolder + "file.txt"
	alreadyExistFile := createTextFile(localDestFolder, "other.txt")
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(unexpectedResultFile)
		_ = os.Remove(expectedOverflowFile)
		_ = os.Remove(alreadyExistFile)
	}()

	// Given
	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Move,
		1,
		localOverflowFolder)

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(unexpectedResultFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedResultFile)
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be present: %s", sourceFile)
	}

	if _, err := os.Stat(expectedOverflowFile); err != nil {
		t.Errorf("File should be present: %s", expectedOverflowFile)
	}
}

// Integration test
func TestMoveToOverflowDirWhileItIsNotEmpty(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localSftpFolder, "file.txt")
	unexpectedResultFile := localDestFolder + "file.txt"
	expectedOverflowFile := localOverflowFolder + "file.txt"
	alreadyExistFile := createTextFile(localOverflowFolder, "other.txt")
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(unexpectedResultFile)
		_ = os.Remove(expectedOverflowFile)
		_ = os.Remove(alreadyExistFile)
	}()

	// Given
	flow := fileflows.NewSFTPFileFlow(
		"Move Nexus files",
		"localhost",
		22,
		sftpPrivateKeyFile,
		remoteInputSftpFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Move,
		1,
		localOverflowFolder)

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(unexpectedResultFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedResultFile)
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be present: %s", sourceFile)
	}

	if _, err := os.Stat(expectedOverflowFile); err != nil {
		t.Errorf("File should be present: %s", expectedOverflowFile)
	}
}

// Integration test
func TestMoveFromLocalDir(t *testing.T) {
	assertFoldersAreEmpty()

	sourceFile := createTextFile(localOverflowFolder, "file.txt")
	expectedResultFile := localDestFolder + "file.txt"
	unexpectedLocalFile := createTextFile(localOverflowFolder, "file.txt")
	defer func() {
		_ = os.Remove(sourceFile)
		_ = os.Remove(expectedResultFile)
		_ = os.Remove(unexpectedLocalFile)
	}()

	// Given
	flow := fileflows.NewLocalFileFlow(
		"Move Nexus files",
		localOverflowFolder,
		".+",
		[]string{localDestFolder},
		fileflows.Move,
		1,
		"")

	// When
	processFlow(flow)

	// Then
	if _, err := os.Stat(unexpectedLocalFile); err == nil {
		t.Errorf("File should not be found: %s", unexpectedLocalFile)
	}

	if _, err := os.Stat(sourceFile); err == nil {
		t.Errorf("File should not be present: %s", sourceFile)
	}

	if _, err := os.Stat(expectedResultFile); err != nil {
		t.Errorf("File should be present: %s", expectedResultFile)
	}
}

func createTextFile(folder string, fileName string) (filePath string) {
	name := folder + fileName
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, _ = file.WriteString("This is a test file.\n")

	log.Printf("Created file: %s", name)

	return name
}

func createGzipFile(folder string, fileName string) (filePath string) {
	name := folder + fileName
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString("This is a test file.\n")
	if err != nil {
		_ = file.Close()
		panic(err)
	}
	_ = file.Close()

	gzipFileName := name + ".gz"
	gz, err := os.Create(gzipFileName)
	if err != nil {
		panic(err)
	}
	gzw := gzip.NewWriter(gz)
	defer gzw.Close()

	_, _ = io.Copy(gzw, file)

	log.Printf("Created file: %s", gzipFileName)

	_ = os.Remove(name)

	return gzipFileName
}
func assertFoldersAreEmpty() {
	if files.CountFiles(localSftpFolder) != 0 {
		log.Fatal("Local sftp folder is not empty")
	}

	if files.CountFiles(localDestFolder) != 0 {
		log.Fatal("Local dest folder is not empty")
	}
}
