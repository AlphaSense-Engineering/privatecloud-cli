# privatecloud-cli

The `privatecloud-cli` is a tool designed to facilitate the management of AlphaSense Enterprise Intelligence Private Cloud environments.

## Prerequisites

- Access to a designated remote Kubernetes cluster on your chosen cloud provider via the `kubectl` CLI tool.
- Ensure you have the necessary permissions to create the following resources in the cluster:
  `ServiceAccount`, `Role`, `RoleBinding`, `ClusterRole`, `ClusterRoleBinding`, and `Pod`.
- Ensure you have the necessary permissions to assign the following permissions to a `Role`:
  - Access to `secrets` with all actions allowed, in the namespaces: `alphasense`, `crossplane`, `mysql`, and `platform`.
  - Access to `pods` with all actions allowed, in the `crossplane` namespace.
  - Access to `pods/log` with all actions allowed, in the `crossplane` namespace.
  - Access to `serviceaccounts` with all actions allowed, in the `crossplane` namespace.
  - Access to `serviceaccounts/token` with all actions allowed, in the `crossplane` namespace.
- Ensure you have the necessary permissions to assign the following permissions to a `ClusterRole`:
  - Access to `storageclasses` in the `storage.k8s.io` group with all actions allowed.
  - Access to `nodes` with all actions allowed.
- If you prefer to compile the project from source, Go v1.23.4 or later must be installed.

## Installation

<!-- ### Pre-compiled Binaries -->

### Compiling from Source

If you prefer to compile the project from source, e.g. for security purposes, follow these steps:

1. Clone the repository:

    ```bash
    git clone https://github.com/AlphaSense-Engineering/privatecloud-cli.git
    cd privatecloud-cli
    ```

2. Build the project:

    ```bash
    go build -o privatecloud-cli
    ```

The binary will be created in the current directory.

## Usage

Currently, the `privatecloud-cli` CLI tool allows you to install the Private Cloud and check the cluster's infrastructure and configuration prior to installation.

See below for instructions on how to use the infrastructure check and installation commands.

### Infrastructure Check Command

The `check` command checks the cluster's infrastructure and configuration prior to installation.

```bash
./privatecloud-cli check <first_step_file>
```

The `<first_step_file>` should be replaced with the path to the first step YAML file in the installation process, such as `step1.yaml`.

### Installation Command

The `install` command installs the AlphaSense Enterprise Kubernetes resources from the specified YAML files.

```bash
./privatecloud-cli install <context> <secrets_file> <first_step_file> <second_step_file> <third_step_file>
```

The `<context>` should be replaced with the name of the Kubernetes context to use for the installation.

The `<secrets_file>` should be replaced with the path to the secrets YAML file in the installation process, such as `init_secrets.yaml`.

The `<first_step_file>`, `<second_step_file>`, and `<third_step_file>` should be replaced with the path to the first, second, and third step YAML files in the
installation process, such as `step1.yaml`, `step2.yaml`, and `step3.yaml`.

## Contributing

While contributions to this repository are generally not expected, AlphaSense Inc. may consider contributions that align with the project's goals and
standards. Any potential contributors should contact AlphaSense Inc. for guidance and permission before proceeding.

If you would like to contribute, please make sure to review both
[LICENSE.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/LICENSE.md) and
[CONTRIBUTING.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/CONTRIBUTING.md) beforehand.

## License

**This publicly accessible Repository is not Open Source.** See [LICENSE.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/LICENSE.md)
for more details.
