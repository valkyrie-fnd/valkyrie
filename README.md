# Valkyrie Game Aggregator

## An open source game aggregator written in Go.
For general information about Valkyrie, please check [this](https://valkyrie.bet/about) out!

Information for developers and other interested parties is found below.

## Installing
**Note**: Windows users are recommended to use PowerShell. Custom tasks may be used for build purposes, e.t.c. To handle that in Windows it's a good idea to be prepared having installed `bash`. This can be done using a package manager (e.g. `Chocolatey`) and install a Linux subsystem (e.g. `WSL`).

Valkyrie is built from source by running:

```shell
go build
```

You can then run Valkyrie using:

```shell
./valkyrie -config path/to/config.yml
```
Two template config files come with the Valkyrie software. These can be found [here](configs/testdata).

### Swagger
To include swagger ui to view all exposed endpoints in Valkyrie you can use the build tag `dev`. After starting Valkyrie, simply direct your web browser to localhost:$<port>$/swagger and take a look at the appropriate endpoints

```shell
go build -tags=dev
```

### Custom tasks

Valkyrie uses [Task](https://taskfile.dev/) as a task runner, e.g. for building the application. To use the tool please refer to its [installation](https://taskfile.dev/installation/) page.

A `Taskfile.yml` is located in the project root which describes custom tasks that can be run for the project.

To list available tasks run:

```shell
task -l
```

TODO: Some tasks cannot be executed on windows?? E.g. task docker:build

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

## Documentation
Check [documentation site](https://valkyrie.bet/docs/) for the latest and most extensive information.

## Performance
Valkyrie aims to be lightweight and fast. While running single instances of Valkyrie and a PAM, we are able to generate and process the following number of bet requests (a.k.a debit/withdraw/stake) for the respective provider implementations (all run locally on an 2021 Macbook Pro w/ M1 MAX CPU, no tracing enabled and INFO log level).


| Provider  | Requests/sec |
|:----------|-------------:|
| Evolution |        40700 |
| Red Tiger |        21000 |
| Caleta    |        29200 |

## Contribution guidelines

Read [contribution guidelines here](./CONTRIBUTING.md)

## Contact

If you have questions regarding an integration you can either post a question [on GitHub](https://github.com/valkyrie-fnd/valkyrie/discussions) or contact us on: help (at) valkyrie.bet
