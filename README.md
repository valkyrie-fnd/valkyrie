# Valkyrie Game Aggregator

An open source game aggregator written in Go.

Valkyrie software presents an easy way to integrate game providers to gaming operators in order to remove the need for additional licensed aggregator software and potentially standardizing the integration. Valkyrie provides a generic interface for operators to integrate with, and have game provider specific modules which can be enabled on a per provider basis.

In it's current shape, Valkyrie is a thread safe, stateless, high performant adapter plug connecting providers and operators.

The software consists of a set of core functions together with a standardized operator interface client and provider specific modules. Valkyrie runs as a service with endpoints for provider wallet transactions, a client communicating with operators' wallet and endpoints for server-to-server game launches for those operators/providers who prefer this game launch flavour. In addition, Valkyrie offers some utilities for front end game client interaction. 

Optional integrations to proprietary operators' protocols can be handled on a case to case basis.

Valkyrie is configurable with respect to providers, operators, logging, tracing and communication timeouts.

For integration testing and kick start purposes, there is an additional project, valkyrie-stubs, avaliable. Valkyrie-stubs contains a mock test bench simulating a boiler plate casino wallet (/genericpam). The valkyrie-stubs wallet publishes services according to Valkyrie OAPI3 PAM client specification, some basic business logic and an in-memory, simple datastore (/memorydatastore). Test benches simulating providers are available in the Valkyrie project itself (/provider/{provider}/test). In other words, Valkyrie project, together with valkyrie-stubs, constitute an environment enabling integration tests of the Valkyrie software in any isolated environment.  

The software is available as a Go binary file, packaged in container or as raw code for anyone to compile and use. Valkyrie is recommended to execute within operators' networks but it can optionally be deployed virtually anywhere.

## Installing

Valkyrie is built from source by running:

```shell
go build
```

You can then run Valkyrie using:

```shell
./valkyrie -config path/to/config.yml
```

### Custom tasks

Valkyrie uses [Task](https://taskfile.dev/) as a task runner, to get the tool please refer to its [installation](https://taskfile.dev/installation/) page.

A `Taskfile.yml` is located in the project root which describes custom tasks that can be run for the project.

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
