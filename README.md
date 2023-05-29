# FileFlow

This Go project provides a utility for moving files from a source location (local or SFTP) to multiple destination folders on a local system. It also allows you to set a maximum limit on the number of files in each destination folder. When a destination folder reaches its maximum capacity, any additional files will be moved to an overflow folder.

## Features

- Move files from a source location (local or SFTP) to multiple destination folders.
- Specify a maximum limit for the number of files in each destination folder.
- Automatically move files to an overflow folder when a destination folder is full.
- Configurable source and destination paths.
- Supports both local file systems and SFTP servers.

## Requirements

- Go 1.20 or above

## Installation

1. Clone the repository:

```shell
git clone https://github.com/chrix75/FileFlow.git
```

2. Change to the project directory:

```shell
cd FileFlow
```

3. Build the project:

```shell
go build
```

## Configuration

The configuration for the utility is stored in a `config.yaml` file. Before running the program, make sure to configure the following settings:

```yaml
delay: 2
file_flows:
  - name: Move ACME files
    server: localhost
    port: 22
    private_key_path: /Users/batman/.ssh/test.sftp.privatekey.file
    from: sftp/acme
    pattern: .+
    to:
      - /Users/Batman/fileflow/acme
    overflow_folder: /Users/Batman/fileflow/overflow
    max_file_count: 3

  - name: Move from ACME overflow
    from: /Users/Batman/fileflow/overflow
    pattern: .+
    to:
      - /Users/Batman/fileflow/acme
    max_file_count: 3
```

Update the `from` with the directory where the source files are located. Adjust the `max_file_count` value to set the maximum number of files that each destination folder can hold. Specify the `overflow_folder` where files will be moved if a destination folder is full.

You can configure multiple destination folders by adding additional entries under the `to` section. 

## Usage

Once you have configured the settings in the `config.yaml` file, run the `FileFlow` executable. The program will start moving files from the source location to the destination folders according to the specified rules.

```shell
./FileFLow config.yaml
```

The program will continuously monitor the source directory for new files. As files are detected, they will be distributed across the destination folders based on the maximum file limit. If a destination folder is full, files will be moved to the overflow folder.

To stop the application, simply press `Ctrl + C` in the terminal.

## Contributing

Contributions are welcome! If you find any issues or would like to suggest enhancements, please open an issue or submit a pull request to the [GitHub repository](https://github.com/chrix75/FileFlow).

## License

This project is licensed under the MIT License.