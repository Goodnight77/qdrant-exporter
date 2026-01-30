from dataclasses import dataclass, field
from typing import Dict, List, Optional


@dataclass(frozen=True)
class MetricSample:
    name: str
    labels: Dict[str, str] = field(default_factory=dict)
    value: float = 0.0


@dataclass
class MetricSeries:
    name: str
    help: Optional[str] = None
    metric_type: Optional[str] = None
    samples: List[MetricSample] = field(default_factory=list)

    def first_value(self) -> float:
        if not self.samples:
            raise ValueError(f"Metric '{self.name}' has no samples")
        return self.samples[0].value
