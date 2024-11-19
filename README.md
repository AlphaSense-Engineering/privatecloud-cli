# privatecloud-installer

The `privatecloud-installer` is a tool designed to facilitate the installation of AlphaSense Enterprise Intelligence Private Cloud environments.

## Prerequisites

- Access to designated remote Kubernetes cluster on your chosen cloud provider via the `kubectl` CLI tool
- If you prefer to compile the project from source, Go v1.22.6 or later must be installed

## Installation

### Pre-compiled Binaries

TBD

### Compiling from Source

If you prefer to compile the project from source, e.g. for security purposes, follow these steps:

1. Clone the repository:

    ```bash
    git clone https://github.com/AlphaSense-Engineering/privatecloud-installer.git
    cd privatecloud-installer
    ```

2. Build the project:

    ```bash
    go build -o privatecloud-installer
    ```

The binary will be created in the current directory.

## Usage

The `privatecloud-installer` CLI tool allows you to install the Private Cloud and check the cluster's infrastructure and configuration prior to installation.

### Infrastructure Check Command

The `check` command checks the cluster's infrastructure and configuration prior to installation.

```bash
./privatecloud-installer check <first_step_file>
```

The `<first_step_file>` should be replaced with the path to the first step YAML file in the installation process, such as `step1.yaml`.

### Installation Command

TBD

## Contributing

While contributions to this project are generally not expected, we appreciate any efforts to improve it.

If you would like to contribute, please make sure to take a look
at [this guideline](https://github.com/AlphaSense-Engineering/privatecloud-installer/blob/main/CONTRIBUTING.md) beforehand.

## License

TBD
