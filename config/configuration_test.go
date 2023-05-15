package config

import (
	"regexp"
	"testing"
)

func TestConfigurationRead(t *testing.T) {
	// Given
	yaml := `
file_flows:
  - name: Move ACME files
    server: localhost
    port: 22 
    from: sftp/acme
    pattern: .+  
    to: /Users/Batman/fileflow/acme
`

	// When
	cfg, err := ReadConfiguration(yaml)
	if err != nil {
		t.Errorf("Error reading configuration: %s", err)
	}

	// Then
	if len(cfg.FileFlows) != 1 {
		t.Errorf("Expected 1 file flow, got %d", len(cfg.FileFlows))
	}

	if cfg.FileFlows[0].Name != "Move ACME files" {
		t.Errorf("Expected Move ACME files, got %s", cfg.FileFlows[0].Name)
	}

	if cfg.FileFlows[0].Server != "localhost" {
		t.Errorf("Expected localhost, got %s", cfg.FileFlows[0].Server)
	}

	if cfg.FileFlows[0].Port != 22 {
		t.Errorf("Expected 22, got %d", cfg.FileFlows[0].Port)
	}

	if cfg.FileFlows[0].SourceFolder != "sftp/acme" {
		t.Errorf("Expected sftp/acme, got %s", cfg.FileFlows[0].SourceFolder)
	}

	if cfg.FileFlows[0].DestinationFolder != "/Users/Batman/fileflow/acme" {
		t.Errorf("Expected /Users/Batman/fileflow/acme, got %s", cfg.FileFlows[0].DestinationFolder)
	}

	if cfg.FileFlows[0].Pattern != ".+" {
		t.Errorf("Expected .+, got %s", cfg.FileFlows[0].Pattern)
	}
}

func TestConfigurationReadWithDefaultValues(t *testing.T) {
	// Given
	yaml := `
file_flows:
  - name: Move ACME files
    server: localhost
    from: sftp/acme
    to: /Users/Batman/fileflow/acme
`

	// When
	cfg, err := ReadConfiguration(yaml)
	if err != nil {
		t.Errorf("Error reading configuration: %s", err)
	}

	// Then
	if len(cfg.FileFlows) != 1 {
		t.Errorf("Expected 1 file flow, got %d", len(cfg.FileFlows))
	}

	if cfg.FileFlows[0].Name != "Move ACME files" {
		t.Errorf("Expected Move ACME files, got %s", cfg.FileFlows[0].Name)
	}

	if cfg.FileFlows[0].Server != "localhost" {
		t.Errorf("Expected localhost, got %s", cfg.FileFlows[0].Server)
	}

	if cfg.FileFlows[0].Port != 22 {
		t.Errorf("Expected 22, got %d", cfg.FileFlows[0].Port)
	}

	if cfg.FileFlows[0].SourceFolder != "sftp/acme" {
		t.Errorf("Expected sftp/acme, got %s", cfg.FileFlows[0].SourceFolder)
	}

	if cfg.FileFlows[0].DestinationFolder != "/Users/Batman/fileflow/acme" {
		t.Errorf("Expected /Users/Batman/fileflow/acme, got %s", cfg.FileFlows[0].DestinationFolder)
	}

	if cfg.FileFlows[0].Pattern != ".+" {
		t.Errorf("Expected .+, got %s", cfg.FileFlows[0].Pattern)
	}
}

func TestDestinationFound(t *testing.T) {
	// Given
	pattern := ".+"
	flow := FileFlow{"Move ACME files", "localhost", 22, "sftp/acme", pattern, "/dest", regexp.MustCompile(pattern)}

	// When
	d := flow.destination("file_A")

	// Then
	if d == "" {
		t.Errorf("Expected destination action, got nothing")
	}

	if d != "/dest/file_A" {
		t.Errorf("Expected /dest/file_A, got %s", d)
	}
}

func TestDestinationNotFound(t *testing.T) {
	// Given
	pattern := "foo_.+"
	flow := FileFlow{"Move ACME files", "localhost", 22, "sftp/acme", pattern, "/dest", regexp.MustCompile(pattern)}

	// When
	d := flow.destination("file_A")

	// Then
	if d != "" {
		t.Errorf("Expected no destination action, got %s", d)
	}
}
