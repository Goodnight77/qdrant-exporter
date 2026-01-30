class ExporterError(Exception):
    """Base error for exporter SDK failures."""


class MetricNotFoundError(ExporterError):
    """Raised when a requested metric is not present in the exporter output."""
