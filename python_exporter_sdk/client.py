from __future__ import annotations

import json
import re
from dataclasses import dataclass
from typing import Dict, Iterable, List, Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

from .exceptions import ExporterError, MetricNotFoundError
from .models import MetricSample, MetricSeries

_METRIC_LINE = re.compile(
    r'^(?P<name>[a-zA-Z_:][a-zA-Z0-9_:]*)'
    r'(?:\{(?P<labels>[^}]*)\})?\s+'
    r'(?P<value>[-+]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][-+]?\d+)?)$'
)


def _parse_labels(raw: str) -> Dict[str, str]:
    if not raw:
        return {}

    labels: Dict[str, str] = {}
    for part in raw.split(","):
        key, value = part.split("=", 1)
        labels[key.strip()] = value.strip().strip('"')
    return labels


def _parse_metrics(text: str) -> Dict[str, MetricSeries]:
    series_map: Dict[str, MetricSeries] = {}

    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line:
            continue
        if line.startswith("# HELP "):
            _, _, name, *help_parts = line.split(" ")
            series = series_map.setdefault(name, MetricSeries(name=name))
            series.help = " ".join(help_parts)
            continue
        if line.startswith("# TYPE "):
            _, _, name, metric_type = line.split(" ", 3)
            series = series_map.setdefault(name, MetricSeries(name=name))
            series.metric_type = metric_type
            continue
        match = _METRIC_LINE.match(line)
        if not match:
            continue

        name = match.group("name")
        labels = _parse_labels(match.group("labels") or "")
        value = float(match.group("value"))
        series = series_map.setdefault(name, MetricSeries(name=name))
        series.samples.append(MetricSample(name=name, labels=labels, value=value))

    return series_map


@dataclass
class ExporterClient:
    base_url: str = "http://localhost:9999"
    timeout: float = 5.0

    def _metrics_url(self) -> str:
        return f"{self.base_url.rstrip('/')}/metrics"

    def fetch_metrics_text(self) -> str:
        request = Request(self._metrics_url(), headers={"Accept": "text/plain"})
        try:
            with urlopen(request, timeout=self.timeout) as response:
                return response.read().decode("utf-8")
        except HTTPError as exc:
            raise ExporterError(f"Exporter returned HTTP {exc.code}") from exc
        except URLError as exc:
            raise ExporterError(f"Failed to reach exporter: {exc.reason}") from exc

    def get_series(self) -> Dict[str, MetricSeries]:
        return _parse_metrics(self.fetch_metrics_text())

    def metric_names(self) -> List[str]:
        return sorted(self.get_series().keys())

    def get_series_by_name(self, name: str) -> MetricSeries:
        series = self.get_series().get(name)
        if series is None:
            raise MetricNotFoundError(f"Metric '{name}' not found")
        return series

    def get_samples(self, name: str) -> List[MetricSample]:
        return self.get_series_by_name(name).samples

    def get_value(self, name: str, labels: Optional[Dict[str, str]] = None) -> float:
        labels = labels or {}
        series = self.get_series_by_name(name)
        for sample in series.samples:
            if all(sample.labels.get(k) == v for k, v in labels.items()):
                return sample.value
        raise MetricNotFoundError(f"Metric '{name}' with labels {labels} not found")

    def to_json(self) -> str:
        payload = {}
        for name, series in self.get_series().items():
            payload[name] = {
                "help": series.help,
                "type": series.metric_type,
                "samples": [
                    {"labels": sample.labels, "value": sample.value}
                    for sample in series.samples
                ],
            }
        return json.dumps(payload, indent=2, sort_keys=True)
