# Valkyrie Game Aggregator
[![](https://img.shields.io/badge/License-MIT%20-brightgreen.svg)](./LICENSE.md) 
[![](https://img.shields.io/github/actions/workflow/status/valkyrie-fnd/valkyrie/gh-workflow.yml)](https://github.com/valkyrie-fnd/valkyrie/actions/workflows/gh-workflow.yml)
![](https://img.shields.io/github/last-commit/valkyrie-fnd/valkyrie)
[![](https://img.shields.io/website?url=https%3A%2F%2Fvalkyrie.bet)](https://valkyrie.bet/docs)
![](https://img.shields.io/github/go-mod/go-version/valkyrie-fnd/valkyrie)
![](https://img.shields.io/github/languages/top/valkyrie-fnd/valkyrie)
![](https://img.shields.io/tokei/lines/github/valkyrie-fnd/valkyrie)
[![](https://img.shields.io/docker/image-size/valkyriefnd/valkyrie)](https://hub.docker.com/r/valkyriefnd/valkyrie)
[![](https://img.shields.io/github/v/release/valkyrie-fnd/valkyrie)](https://github.com/valkyrie-fnd/valkyrie/releases)
![](https://img.shields.io/maintenance/yes/2023)
## An open source game aggregator written in Go.
For general information about Valkyrie, please check [this](https://valkyrie.bet/about) out!

Information for developers and other interested parties is found below.

## Building

Valkyrie is built from source by running:

```shell
go build
```

## Running

You can then run Valkyrie using:

```shell
./valkyrie -config path/to/config.yml
```
Two template config files come with the Valkyrie software. These can be found [here](configs/testdata).

### Custom tasks

Valkyrie uses [Task](https://taskfile.dev/) as a task runner, e.g. for building the application.
To use the tool please refer to its [installation](https://taskfile.dev/installation/) page.

> **Note**: Windows users are recommended to use PowerShell. Custom tasks may be used for build purposes, etc.
To run most of these in Windows you need to have installed [wsl](https://www.microsoft.com/store/productId/9P9TQF7MRM4R)
to enable `bash`. Another helpful tool is [Chocolatey](https://chocolatey.org/), which can be used on windows to
install other programs, like Task.

A `Taskfile.yml` is located in the project root which describes custom tasks that can be run for the project.

To run the default tasks (generate, lint & test), simply run:
```shell
task
```

To list available tasks run:

```shell
task -l
```

### Docker

Valkyrie image is built using:

```shell
docker build -t valkyrie .
```

You can then run the Valkyrie container using:

```shell
docker run -v /absolute/path/config.yml:/app/config.yml valkyrie -config config.yml
```

### Helm

A Helm chart is provided to run Valkyrie in Kubernetes.

For instructions how to use the chart please refer to [helm/README.md](./helm/README.md).

### Linux service

Valkyrie can be installed as `systemd` service by these steps:
- obtaining a linux [release package of valkyrie](/releases/latest/), named like `valkyrie-X.X.X-linux-amd64.tar.gz`
- unpacking and setting up a configuration file named `config.yml`
- executing `./svc.sh install` to setup the service

### Swagger

To include swagger ui to view all exposed endpoints in Valkyrie you can use the build tag `dev`.
After starting Valkyrie, simply direct your web browser to localhost:<port>/swagger and take a look at the
appropriate endpoints.

```shell
go build -tags=dev
```

## Documentation

Check [documentation site](https://valkyrie.bet/docs/) for the latest and most extensive information.

### Architecture
You can view the project following the c4 model by starting a [structurizr](https://structurizr.com/) server on [localhost:8090](http://localhost:8090) using task:
```shell
task c4model
```
System landscape view where Valkyrie fits in:

![system landscape](/structurizr/structurizr-1-OnlineGaming.png)

Components view of Valkyrie, how the internals connect to each other.
![components view](/structurizr/structurizr-1-Valkyrie-Components.png)

### Codebase structure
Going through the code there are two folders that are more important than others, `<providers>` and `<pam>`. In these you will find the relevant components that makes Valkyrie modular.
In `<providers>` you will find each available game provider module that is integrated with Valkyrie. More in detail how to create a new provider module and the code structure of them can be found [here](./example/README.md).
When starting a Valkyrie instance you provide it with a config yaml file containing what provider modules you want to have enabled. Read more about the configuration file [here](https://valkyrie.bet/docs/get-started/configuration).
In the root `<provider>`-folder is the implementations of `provider/docs/operator_api.yml`. It is the specification the operator can use to send request to Valkyrie that will be forwarded toward the provider. You can read more [here](https://valkyrie.bet/docs/operator/valkyrie-operator-api).

`<pam>` contains implemented PAMs that are available. Each provider is using `pam.PamClient` found in `/pam/client.go` to communicate with the PAM. The available PAMs should implement this interface. Similar to the providers, what PAM is used is based on the configuration file provided at runtime. There is an oapi specification that is used by `<genericpam>` called `pam_api.yml` that can be read [here](https://valkyrie.bet/docs/wallet/valkyrie-pam/valkyrie-pam-api).
Fulfilling this specification will enable you to use `<genericpam>`. In the diagram above, `<genericpam>` and `<vplugin>` would be the "PAM Client".
"PAM Plugin", from the diagram, would be used in the case of `<vplugin>`. `<vplugin>` is utilizing hashicorp go-plugin to keep the pam implementation in a separate process from Valkyrie. This enables you to keep your own private implementation of a PAM. More on this can be read [here](https://valkyrie.bet/docs/wallet/vplugin/vplugin-introduction).
``` text
+ configs/                # The parsing of configuration yaml
+ example/                # Code examples
  + example-game-provider # An example provider module with information about the different components needed to make your own provider module
+ internal/               # Shared utility packages
+ ops/                    # Logging and tracing
+ pam/
  + generic_pam/          # PAM following pam_api.yaml specification
  + vplugin/              # PAM plugin
    - client.go           # Contains the interface the plugin need to follow
  - generate.go           # used to generate the models defined in pam_api.yml, using oapi-codegen
  - client.go             # Contains interface the pam must fulfil. The interface that is used by the game provider modules.
+ provider/
  + gameProvider_X/       # Game provider modules
  + gameProvider_Y/
  + evolution/
  + redtiger/
  - xyz_controller.go     # Handler functions for operator_api.yml
+ rest/                   # Http client for making requests
+ routes/                 # Setting up the routes for each provider and the operator
+ server/                 # Starting the Valkyrie servers, exposed toward Operator and Providers
```
"Provider Server" and "Operator Server" from the Component diagram above would be in `routes/routes.go` where the operator and provider routes are setup. It utilizes `provider/registry.go` to add the routes to the `fiber.App`. Each provider needs to implement both "Operator Router" and "Provider Router". They are implemented under `{gameProvider}/router.go`. For more details on all the "Provider Module"-components the [example game provider](./example/README.md) can be viewed, with documentation through out the example.

### CI

Valkyrie uses GitHub actions for continuous integration, which is described under [workflows](./.github/workflows).

Pull requests will trigger Lint and Test jobs which are expected to pass before merging.

Commits to main branch will rerun Lint and Test, and if successful it will also build a pre-release
version (`x.y.z-pre.revno`), which is uploaded to Docker Hub.

### Releasing

1. Choose a suitable release version following the [Semantic Versioning](https://semver.org/spec/v2.0.0.html) format
2. Before creating the release, make sure to update [CHANGELOG.md](./CHANGELOG.md), moving the entries from `Unreleased`
   to your release version and describe any new notable changes since the previous release.
3. Finally, trigger the `Release` workflow in GitHub Actions and specify the version as argument.

The GitHub actions workflow will perform the following steps:
* Run the full CI suite
* Build and push container to Docker Hub
* Build and push helm chart to Docker Hub
* Build binaries for various platforms and upload to GitHub releases
* Tag the version in git

## Performance
Valkyrie aims to be lightweight and fast. While running single instances of Valkyrie and a PAM, we are able to generate and process the following number of bet requests (a.k.a debit/withdraw/stake) for the respective provider implementations (all run locally on an 2021 MacBook Pro w/ M1 MAX CPU, no tracing enabled and INFO log level).


| Provider  | Requests/sec |
|:----------|-------------:|
| Evolution |        40700 |
| Red Tiger |        21000 |
| Caleta    |        29200 |

## Contribution guidelines

Read [contribution guidelines here](./CONTRIBUTING.md)

## Contact

If you have questions regarding an integration you can either post a question [on GitHub](https://github.com/valkyrie-fnd/valkyrie/discussions) or contact us on: help (at) valkyrie.bet
