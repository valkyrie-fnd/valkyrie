# Releasing

## Update CHANGELOG

Before creating a release, make sure to update [CHANGELOG.md](../CHANGELOG.md).

* Choose a suitable release version following the [Semantic Versioning](https://semver.org/spec/v2.0.0.html) format.
* Move the entries from `Unreleased` to your release version.

Finally, trigger the `Release` workflow in GitHub Actions and specify the version as argument.

The workflow will perform the following actions:
* Run the full CI suite
* Build and push docker container
* Build and push helm chart
* Build binaries for various platforms and upload to GitHub
* Tag the version in git
