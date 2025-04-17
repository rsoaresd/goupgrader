# goupgrader

goupgrader is a tool designed to simplify managing and upgrading Go modules in your Go projects. It helps you to automatically upgrade specific dependencies in your `go.mod` file to specified versions or branches.

Additionally, it includes a command to generate a config file tailored to a specific OpenShift release. By analyzing which Kubernetes version that OpenShift version uses, it matches it with a compatible operator-sdk version and collects the relevant dependency versions. This config can then be used to upgrade your project in alignment with the selected OpenShift release.

# Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Testing](#testing)
- [Contributing](#contributing)

## Installation
To install the `goupgrader` binary, follow these steps:

1. Clone the latest version of the repository:
   ```sh
   git clone https://github.com/rsoaresd/goupgrader.git
   cd goupgrader
   ````
2. Run the following command to install the binary:
   ```sh
   make install
   ````

## Usage

### Upgrade dependencies
To upgrade dependencies in your Go project, you must provide the `--config` and `--project` flags. Here's how to use it:

### Example
```sh
goupgrader upgrade --config <config-path> --project <your-go-project-path>
```

### Generate config dependencies based on a Openshift version
Generates a YAML configuration file for upgrading Go project dependencies based on the Kubernetes version used by a specific OpenShift version.

```sh
goupgrader generate --target-openshift-version=<openshift-version> --in-use-op-sdk-version=<operator-sdk-version> --output=<config-file-path>

```

## Configuration
You must provide a YAML configuration file that lists the dependencies you want to upgrade. Each dependency can specify either a version or a branch, but not both.

### Example
```sh
dependencies:
  - package: "sigs.k8s.io/controller-runtime"
    version: "v0.19.3"
  - package: "github.com/openshift/api"
    branch: "release-4.18"
```

### Configuration Fields
- **`dependencies`**: A list of dependencies to upgrade.
  - **`package`** (`string`, required): The import path of the Go module to upgrade.
  - **`version`** (`string`, optional): A semantic version to upgrade the module to (e.g., `"v1.2.3"`). Cannot be used with `branch`.
  - **`branch`** (`string`, optional): A Git branch to track. The latest commit hash from this branch will be fetched and used as a pseudo-version. Cannot be used with `version`.


## Testing
To run all unit tests, you can simply use:
```sh
make test
```

## Contributing
Thank you so much for considering contributing to this project!

If you would like to contribute, please follow these instructions:
1. Fork the `goupgrader` repository.
2. Install Go version 1.22 or newer from [golang.org/dl](https://golang.org/dl/).
3. Locally, clone your fork (`git clone https://github.com/<your-gh-username>/goupgrader.git`).
4. Create a new branch for local development (`git checkout -b <branch-name>`).
5. Work on your changes locally.
6. Review and test your changes:
    ```sh
    golangci-lint run
    ```
    ```sh
    make test
    ```
3. Commit your changes (`git commit -am 'Add feature'`).
4. Push your changes to your branch (`git push origin feature/feature-name`).
5. Open a Pull Request.

Note that your changes should follow the existing code style and include tests when relevant.
