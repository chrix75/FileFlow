// Package fileflows provides flows configuration.
package fileflows

import (
	"gopkg.in/yaml.v3"
	"log"
	"regexp"
)

type FlowOperation int

const (
	Move = iota
	Compression
	Decompression
)

// FFConfig is the presentation of all flows defined in the config YAML file.
type FFConfig struct {
	FileFlows []FileFlow `yaml:"file_flows"`
}

// FileFlow represents a flow defined in the config YAML file.
type FileFlow struct {
	Name               string
	Server             string
	Port               int
	PrivateKeyPath     string `yaml:"private_key_path"`
	SourceFolder       string `yaml:"from"`
	Pattern            string
	DestinationFolders []string `yaml:"to"`
	Regexp             *regexp.Regexp
	Operation          FlowOperation
	MaxFileCount       int
	OverflowFolder     string
}

// ReadConfiguration reads a config YAML and returns a FFConfig struct.
func ReadConfiguration(cfg string) (*FFConfig, error) {
	read := FFConfig{}
	err := yaml.Unmarshal([]byte(cfg), &read)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG - Read configuration: %+v", read)

	flows := make([]FileFlow, len(read.FileFlows))
	for i, flow := range read.FileFlows {
		pattern := usedPattern(&flow)

		if isSFTPFlow(&flow) {
			flows[i] = NewSFTPFileFlow(
				flow.Name,
				flow.Server,
				usedPort(&flow),
				flow.PrivateKeyPath,
				flow.SourceFolder,
				pattern,
				flow.DestinationFolders,
				flow.Operation,
				flow.MaxFileCount,
				flow.OverflowFolder)
		} else {
			flows[i] = NewLocalFileFlow(
				flow.Name,
				flow.SourceFolder,
				pattern,
				flow.DestinationFolders,
				flow.Operation,
				flow.MaxFileCount,
				flow.OverflowFolder)
		}
	}

	result := FFConfig{
		flows,
	}

	log.Printf("DEBUG - Used configuration: %+v", result)

	return &result, nil
}

func isSFTPFlow(f *FileFlow) bool {
	return f.Server != "" && f.PrivateKeyPath != ""
}

func usedPattern(f *FileFlow) string {
	var usedPattern string
	if f.Pattern == "" {
		usedPattern = ".+"
	} else {
		usedPattern = f.Pattern
	}

	return usedPattern

}

func usedPort(f *FileFlow) int {
	var usedPort int
	if f.Server != "" && f.Port == 0 {
		usedPort = 22
	} else {
		usedPort = f.Port
	}
	return usedPort
}

func (f *FileFlow) destination(path string) string {
	if f.Regexp.MatchString(path) {
		return f.DestinationFolders[0] + "/" + path
	}
	return ""
}

func (f *FileFlow) IsRemote() bool {
	return f.Port > 0
}

// NewSFTPFileFlow creates a new SFTP file flow.
// name is the name of the flow.
// privateKeyPath is the path to the private key file to SFTP connection.
// sourceFolder is the path to the source folder from where files are downloaded. This path is relative to the SFTP user root folder.
// pattern is the regular expression used to match files.
// destinationFolders is the list of destination folders where files are downloaded.
// operation is the operation to perform on the files: MOVE, COMPRESSION, DECOMPRESSION. Format compression supported is gzip.
// maxFileCount is the maximum number of files that can be downloaded to a destination folder.
// overflowFolder is the path to the folder where overflow files are stored.
//
// For overflow, it's valid when one destination folder and maxFileCount are specified.
// By default, when the number of files reaches the maxFileCount, file downloads are stopped.
// But, if overflowFolder is specified, file are downloaded into the overflow folder.
func NewSFTPFileFlow(name string,
	server string, port int,
	privateKeyPath string,
	sourceFolder, pattern string,
	destinations []string,
	operation FlowOperation,
	maxFileCount int,
	overflowFolder string) FileFlow {

	if server == "" || port == 0 || privateKeyPath == "" {
		log.Fatal("SFTP flow configuration error: server, port or private_key_path is empty")
	}

	return FileFlow{
		name,
		server,
		port,
		privateKeyPath,
		sourceFolder,
		pattern,
		destinations,
		regexp.MustCompile(pattern),
		operation,
		maxFileCount,
		overflowFolder,
	}
}

// NewLocalFileFlow creates a new local file flow.
// See NewSFTPFileFlow for details.
func NewLocalFileFlow(name, sourceFolder, pattern string,
	destinations []string,
	operation FlowOperation,
	maxFileCount int,
	overflowFolder string) FileFlow {

	if len(destinations) > 1 && overflowFolder != "" {
		log.Fatal("Overflow folder cannot be specified with multiple destinations")
	}

	return FileFlow{
		Name:               name,
		SourceFolder:       sourceFolder,
		Pattern:            pattern,
		DestinationFolders: destinations,
		Regexp:             regexp.MustCompile(pattern),
		Operation:          operation,
		MaxFileCount:       maxFileCount,
		OverflowFolder:     overflowFolder,
	}
}
