# Python Exporter SDK

This package is a thin Python client for the Go exporter service.
It reads the exporter's `/metrics` endpoint and parses Prometheus text dynamically.

## What it is for

- read exporter metrics from Python
- inspect metric names and values
- query labeled collection metrics
- avoid changing the SDK every time the Go exporter adds a new metric

## What you need first

Before using the SDK, make sure the Go exporter is running.

### Local stack

From the `simple/` folder:

```bash
set -a
source .env.local
set +a
docker compose --profile local up -d
```

### Cloud stack

From the `simple/` folder:

```bash
set -a
source .env.cloud
set +a
docker compose up -d
```

### Direct Go run

If you want to run the exporter without Docker:

```bash
set -a
source .env.cloud
set +a
go run .
```

## Install for local development

```bash
python -m pip install -r requirements.txt
python -m pip install -e .
```

## Test the SDK

Run the unit test:

```bash
python -m pytest tests/test_python_exporter_sdk.py
```

Run all Python tests:

```bash
python -m pytest tests
```

Run the live smoke test against a running exporter:

```bash
set EXPORTER_URL=http://localhost:9999
python -m pytest tests/test_live_exporter_sdk.py -m integration
```

Run the plain inspection script:

```bash
set EXPORTER_URL=http://localhost:9999
python tests/inspect_exporter.py
```

## Example usage

```python
from python_exporter_sdk import ExporterClient

client = ExporterClient("http://localhost:9999")

print(client.metric_names())
print(client.get_value("qdrant_up"))
print(client.get_value("qdrant_collection_points", {"collection": "aaaaa"}))
print(client.get_samples("qdrant_collection_points"))
```

## Behavior

- `metric_names()` returns every metric the exporter currently exposes
- `get_value()` can read plain metrics or labeled metrics
- new Go metrics should appear automatically without SDK code changes
