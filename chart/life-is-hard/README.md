# life-is-hard Helm Chart

This chart deploys the `life-is-hard` backend service along with Postgres and Redis using Bitnami subcharts.

## Configuration

All configurable parameters are documented in `values.yaml`. Defaults target a production setup with persistence enabled for both Postgres and Redis and two application replicas.

Install the chart with:

```bash
helm install my-release ./chart/life-is-hard
```

