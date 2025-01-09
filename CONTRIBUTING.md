# Contributing

While contributions to the Software are generally not expected, AlphaSense Inc. may consider contributions that align with the project's goals and
standards. Any potential contributors should contact AlphaSense Inc. for guidance and permission before proceeding. Contributions to the Software would
be subject to the same terms and conditions as the rest of the Software, and the rights to the contributed Software would be transferred to AlphaSense Inc.

**This publicly accessible repository is not Open Source.** See [LICENSE.md](https://github.com/AlphaSense-Engineering/privatecloud-cli/blob/main/LICENSE.md)
for more details.

The information below is provided for informational purposes only. It is not an approval for any contributions.

## Local Development

### Prerequisites

- Docker
- Go v1.23.4 or later
- Task aka Taskfile

### Setting Up

1. Clone the repository:

    ```bash
    git clone https://github.com/AlphaSense-Engineering/privatecloud-cli.git
    cd privatecloud-cli
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

### Available Commands

We use **Task** to manage development commands. The following commands are available:

- **Enabling Git Hooks:**

  ```bash
  task githooks
  ```

- **CI Commands:**
  - **Full Pipeline:**
  
    ```bash
    task ci
    ```

  - **Linting:**
  
    ```bash
    task ci-lint
    ```

  - **Testing:**
  
    ```bash
    task ci-test
    ```

- **Building the Application:**

  ```bash
  task build
  ```

- **Running the Application:**
  
  ```bash
  task run -- <command> <args>
  ```

  - **Running the Check Command:**
  
    ```bash
    task check -- <first_step_file>
    ```

    **N.B.** See available flags by running `task check -- -h`.

  - **Running the Install Command:**
  
    ```bash
    task install -- <context> <secrets_file> <first_step_file> <second_step_file> <third_step_file>
    ```

    **N.B.** See available flags by running `task install -- -h`.

  - **Running the Pod Command:**
  
    ```bash
    task pod
    ```

    This command requires `ENVCONFIG` environment variable to be set with the base64 encoded YAML representation of the `EnvConfig` Kubernetes resource.

- **Building the Pod Image:**
  
  ```bash
  task pod-image-build
  ```

## Architecture

### Infrastructure Check Command

The `check` command is designed to ensure that the infrastructure in your cloud environment is ready for deployment. It performs several steps to achieve this,
including spinning up a **Pod** in the Kubernetes cluster and using a Docker image with the same binary but running the `pod` command, which acts as a hidden
mode.

Here is a high-level overview of the process:

1. The command first ensures that the necessary `crossplane` namespace exists in the cluster. If it does not, it creates it.

2. The command creates the required **ServiceAccount**, **Role**, **ClusterRole**, **RoleBinding**, and **ClusterRoleBinding** resources in the cluster.
These resources define the permissions needed for the **Pod** to operate correctly.

3. The command then creates a **Pod** in the cluster. This **Pod** uses a Docker image that contains the same binary as the `privatecloud-cli` but runs
the `pod` command, specifically designed for the in-cluster check operation.

4. The **Pod** is started, and the command waits for it to reach the running state.

5. Once the **Pod** is running, the command streams the logs from the **Pod** to provide real-time feedback on the in-cluster check operation.

6. After the in-cluster check operation is complete, the command cleans up all of the resources it created.

The `check` command uses a **Pod** and a Docker image to utilize the same codebase while isolating the in-cluster check operation within a controlled
environment in the Kubernetes cluster. This method ensures that the in-cluster check operation has the required permissions and environment to execute
its tasks without needing any additional commands or permissions that would be necessary if run locally.

### Installation Command

The `install` command is designed to install the AlphaSense Enterprise Kubernetes resources from the specified YAML files.

The command performs the following steps:

1. The command runs the infrastructure check to ensure the cluster is ready for installation.

2. The command sequentially applies the YAML files to the cluster. Step 1 file is applied twice to account for the resource mapping not being found on the first
apply.

3. The command waits for the phases to be completed.

## Styleguides

This project does not possess any specific styleguides. Please, make sure to follow general best practices used in Go community.

You can use the following resources for the reference:

- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Effective Go](https://go.dev/doc/effective_go)

### Commit Messages

This project adheres to the [Conventional Commits](https://conventionalcommits.org/en/v1.0.0/) specification. Please, make sure that your commit messages
follow that specification.
