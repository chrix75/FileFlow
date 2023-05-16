// Package fileflows provides flows configuration.
package fileflows

import (
	"gopkg.in/yaml.v3"
	"log"
	"regexp"
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
	SourceFolder       string `yaml:"from"`
	Pattern            string
	DestinationFolders []string `yaml:"to"`
	Regexp             *regexp.Regexp
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
		flows[i] = FileFlow{
			flow.Name,
			flow.Server,
			usedPort(&flow),
			flow.SourceFolder,
			pattern,
			flow.DestinationFolders,
			regexp.MustCompile(pattern),
		}
	}

	result := FFConfig{
		flows,
	}

	log.Printf("DEBUG - Used configuration: %+v", result)

	return &result, nil
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
	if f.Port == 0 {
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
func NewFileFlow(name string, server string, port int, sourceFolder string, pattern string, destinations []string) FileFlow {
	return FileFlow{
		name,
		server,
		port,
		sourceFolder,
		pattern,
		destinations,
		regexp.MustCompile(pattern),
	}
}
