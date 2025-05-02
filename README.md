# privatecloud-cli

The `privatecloud-cli` is a tool designed to facilitate the management of AlphaSense Enterprise Intelligence Private Cloud environments.

For simplicity, AlphaSense Enterprise Intelligence Private Cloud will be referred to as Private Cloud.

## Prerequisites

- Access to a designated remote Kubernetes cluster on your chosen cloud provider via the [`kubectl`](https://kubernetes.io/docs/reference/kubectl) CLI tool.
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
- If you prefer to compile the project from source, [Go](https://go.dev) v1.24.2 or later must be installed.

## Compatibility

Latest `privatecloud-cli` version is generally compatible with the latest Private Cloud version.

Refer to the compatibility matrix below to determine the appropriate `privatecloud-cli` version for your specific Private Cloud version:

| `privatecloud-cli` Version | Private Cloud Version |
|----------------------------|-----------------------|
| v0.1.0 and above           | v2.0.1 and above      |

## Installation

### Pre-compiled Binaries

Pre-compiled binaries are available for download from the [Releases](https://github.com/AlphaSense-Engineering/privatecloud-cli/releases) page.

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

When compiling from source and running the tool, you need to manually set the Docker image for the Private Cloud CLI Pod using the `--docker-image` flag.
If you do not specify this flag, the tool will default to the `privatecloud-cli-pod:dev` image, which is not available in the default Docker repository.

## Usage

Currently, the `privatecloud-cli` CLI tool allows you to install the Private Cloud and check the cluster's infrastructure and configuration prior to
installation.

See below for instructions on how to use the infrastructure check and installation commands.

### Infrastructure Check Command

The `check` command checks the cluster's infrastructure and configuration prior to installation.

```bash
./privatecloud-cli check <first_step_file>
```

The `<first_step_file>` should be replaced with the path to the first step YAML file in the installation process, such as `step1.yaml`.

### Installation Command

The `install` command installs the Private Cloud Kubernetes resources from the specified YAML files.

```bash
./privatecloud-cli install <context> [<secrets_file>] <first_step_file> <second_step_file> <third_step_file>
```

The `<context>` should be replaced with the name of the Kubernetes context to use for the installation. You can get the current context by running
`kubectl config current-context`.

The `[<secrets_file>]` is optional and could be replaced with the path to the secrets YAML file in the installation process, such as `init_secrets.yaml`.

The `<first_step_file>`, `<second_step_file>`, and `<third_step_file>` should be replaced with the path to the first, second, and third step YAML files in the
installation process, such as `step1.yaml`, `step2.yaml`, and `step3.yaml`.

## Contributing

While contributions to this project are generally not expected, we appreciate any efforts to improve it.

If you would like to contribute, please make sure to review both
[LICENSE.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/LICENSE.md) and
[CONTRIBUTING.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/CONTRIBUTING.md) beforehand.

## License

**This repository is not Open Source.** See [LICENSE.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/LICENSE.md)
for more details.
