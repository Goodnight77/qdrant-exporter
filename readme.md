# qdrant-exporter

A simple Prometheus exporter for Qdrant vector database that exposes per-collection metrics.

## What it does

- scrapes the Qdrant API for collection data
- exposes exporter metrics at `/metrics`
- works with Prometheus and Grafana

## Supported Qdrant modes

- local Qdrant in Docker: supported
- local Qdrant running on your host: supported
- Qdrant Cloud: supported if `QDRANT_URL` and `QDRANT_API_KEY` are set

The app reads:

- `QDRANT_URL`
- `QDRANT_API_KEY`

If `QDRANT_API_KEY` is present, the exporter sends it as the `api-key` header.

## How to run

### Local Qdrant mode with Docker

From the `qdrant-exporter/` folder:

```bash
docker compose --env-file .env.local --profile local up -d
```

Use this mode when you want the exporter to talk to the local Qdrant container.

This starts:

- `qdrant`
- `exporter`
- `prometheus`
- `grafana`

### Qdrant Cloud mode with Docker

Then start the stack without the local Qdrant profile:

```bash
docker compose --env-file .env.cloud up -d
```

This mode uses the cloud values from `.env.cloud` and does not start the local Qdrant container.

This starts:

- `exporter`
- `prometheus`
- `grafana`

### Run the exporter against Qdrant Cloud

If you want to use Qdrant Cloud, set these values in `.env` or your shell before starting the exporter:

- `QDRANT_URL` to your Qdrant Cloud endpoint
- `QDRANT_API_KEY` to your Qdrant Cloud API key

```bash
set QDRANT_URL=https://your-cluster-url
set QDRANT_API_KEY=your_api_key
go run .
```

If you run the Docker stack instead, make sure the exporter container also receives those same variables.

If you want to use the cloud setup, do not enable the `local` profile. That keeps the Qdrant container out of the stack.

## Metrics endpoint

Visit:

- [Exporter metrics](http://localhost:9999/metrics)

Example metrics:

```text
# HELP qdrant_up Whether Qdrant is reachable
# TYPE qdrant_up gauge
qdrant_up 1
```

## Prometheus ports

These ports are different because there are two Prometheus views:

- `http://localhost:9091` is the **host-mapped Prometheus UI** from Docker Compose
- `http://localhost:9090` is the **container’s internal Prometheus port**

In this project, you should use:

- [Prometheus graph](http://localhost:9091/graph)
- [Prometheus targets](http://localhost:9091/targets)
- [Prometheus metrics](http://localhost:9091/metrics)

Why you may not see the new Qdrant metrics immediately:

- `[Prometheus metrics](http://localhost:9091/metrics)` shows Prometheus' own internal metrics, not the exporter’s custom `qdrant_*` metrics
- the exporter’s custom metrics are exposed at [http://localhost:9999/metrics](http://localhost:9999/metrics)
- Prometheus only stores what it scrapes from the exporter, so query the metric names in [Prometheus graph](http://localhost:9091/graph) after the scrape runs
- if the exporter is not reachable from Prometheus, check [Prometheus targets](http://localhost:9091/targets) first

If Prometheus is running in Docker Compose, it should scrape `exporter:9999`, not `localhost:9999`.
`localhost` inside the Prometheus container means the Prometheus container itself, not the exporter container.
Use `http://localhost:9999/metrics` in your browser, but Prometheus must use `http://exporter:9999/metrics` on the Docker network.

If you want to confirm Prometheus is scraping the exporter correctly, check:

- [Prometheus targets](http://localhost:9091/targets)

## Qdrant data persistence

- Qdrant data is stored in the named Docker volume `qdrant-storage`
- collections survive `docker compose up -d` and `docker compose down`
- collections are lost if the volume is deleted, for example with `docker compose down -v`
- the local Qdrant container only starts when you use `docker compose --env-file .env.local --profile local up -d`

## Next project step

From `queue.md`, the next necessary step is:

- **Task 6: Add Health Endpoint**

That is the next unfinished core task after the current exporter, client, collector, and compose setup.

## Python SDK

A Python SDK can be added after the exporter behavior is stable and the API surface stops changing.

Best time to build it:

- after the current Docker/local/cloud workflow is confirmed
- after the health endpoint and any remaining core metrics are settled
- before or alongside Grafana/dashboard packaging if you want a user-facing integration layer

## Useful commands

```bash
docker compose ps
docker compose logs qdrant
docker compose logs exporter
```



## troubleshooting (important)
### local vs cloud confusion
if local didnt work check those it might that the env still have the qdrant credentials
```shell
docker compose logs exporter
docker compose exec exporter sh -lc 'echo $QDRANT_URL'
docker compose exec exporter sh -lc 'wget -qO- http://qdrant:6333/collections'
```

#### Expected for local mode:

  - QDRANT_URL should be http://qdrant:6333
  - wget .../collections should return JSON or at least connect

### docker compose down didnt work
```shell
# if you got smthg like ! Network simple-net  Resource is still in use                                              0.0s 

docker compose ps
docker stop NameOfTheContainer
docker compose down
```

