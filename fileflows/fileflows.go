package fileflows

import (
	"gopkg.in/yaml.v3"
	"log"
	"regexp"
)

type FFConfig struct {
	FileFlows []FileFlow `yaml:"file_flows"`
}

type FileFlow struct {
	Name               string
	Server             string
	Port               int      `default:"22"`
	SourceFolder       string   `yaml:"from"`
	Pattern            string   `default:".+"`
	DestinationFolders []string `yaml:"to"`
	Regexp             *regexp.Regexp
}

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
