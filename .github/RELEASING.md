# Releasing

## Update CHANGELOG

Before creating a release, make sure to update [CHANGELOG.md](../CHANGELOG.md).

* Choose a suitable release version following the [Semantic Versioning](https://semver.org/spec/v2.0.0.html) format.
* Move the entries from `Unreleased` to your release version.

## Tag the release

The release process is triggered by pushing a git tag containing the version (`vX.X.X`-format).

Sanity check, make sure you are standing in correct branch with expected revision (most likely `main`).

You may first list the existing tags (again, sanity check):

```shell
git tag -l
```

Tag the revision with a well-formed version:

```shell
git tag -a vX.X.X -m "Helpful comment"
```

Push the tag to trigger the release workflow:

```shell
git push origin vX.X.X
```

Our configured GitHub Actions should take care of the rest and publish to appropriate repositories.
