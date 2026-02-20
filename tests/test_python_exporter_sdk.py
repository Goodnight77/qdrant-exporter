from python_exporter.client import _parse_metrics


def test_parse_metrics_extracts_samples_and_labels():
    text = """# HELP qdrant_up Whether Qdrant is reachable
    # TYPE qdrant_up gauge
    qdrant_up 1
    # HELP qdrant_collection_points number of points in the collection
    # TYPE qdrant_collection_points gauge
    qdrant_collection_points{collection="collection_name1"} 0
"""

    series_map = _parse_metrics(text)

    assert "qdrant_up" in series_map
    assert series_map["qdrant_up"].metric_type == "gauge"
    assert series_map["qdrant_up"].samples[0].value == 1

    points = series_map["qdrant_collection_points"]
    assert points.samples[0].labels == {"collection": "collection_name1"}
    assert points.samples[0].value == 0
