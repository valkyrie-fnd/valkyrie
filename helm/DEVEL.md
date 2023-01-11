# Development guide

Development guide for this helm chart.

## Packaging

This helm chart is packaged and pushed automatically to [docker hub](https://hub.docker.com/r/valkyriefnd/valkyrie-chart/tags) using CI tooling,
which uses predefined task commands (see `helm-push` in _Taskfile.yml_).

### Manual steps

You may also manually package the chart by following these steps:

Package the chart:

```shell
helm package ./
```

Push the packaged chart to repo:

```shell
helm push valkyrie-chart-x.y.z.tgz oci://docker.io/valkyriefnd/
```

You can now access the chart using the OCI registry:

```shell
helm show readme oci://docker.io/valkyriefnd/valkyrie-chart
```

## Troubleshooting

Render the k8s resources produced by the chart:

```shell
helm template ./ --debug
```

This is useful to see what the templates actually produce.
