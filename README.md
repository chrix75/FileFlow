# FileFlow

## Configuration file format
The FileFLow configuration file is written in YAML format.

### Simple configuration
```yaml
file_flows:
  - name: Move ACME files
    server: localhost
    port: 22 
    from: sftp/acme
    pattern: .+  
    to: /Users/Batman/fileflow/acme
```
