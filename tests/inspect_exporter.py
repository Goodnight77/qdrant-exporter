import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python_exporter import ExporterClient


def main() -> None:
    exporter_url = os.getenv("EXPORTER_URL", "http://localhost:9999")
    client = ExporterClient(exporter_url, timeout=15.0)

    print(f"Exporter URL: {exporter_url}")

    metric_names = client.metric_names()
    print("\nMetric names:")
    for name in metric_names:
        print(f"- {name}")

    try:
        qdrant_up = client.get_value("qdrant_up")
        print(f"\nqdrant_up: {qdrant_up}")
    except Exception as exc:
        print(f"\nqdrant_up: error: {exc}")

    try:
        points_samples = client.get_samples("qdrant_collection_points")
        collection_names = sorted(
            {
                sample.labels.get("collection", "")
                for sample in points_samples
                if sample.labels.get("collection")
            }
        )

        print(f"\nCollections found: {len(collection_names)}")
        print(f"Showing top 5 collections:\n")

        for name in collection_names[:5]:
            points = client.get_value("qdrant_collection_points", {"collection": name})
            vectors = client.get_value("qdrant_collection_vectors", {"collection": name})
            indexed = client.get_value(
                "qdrant_collection_indexed_vectors", {"collection": name}
            )
            segments = client.get_value("qdrant_collection_segments", {"collection": name})
            status = client.get_value("qdrant_collection_status", {"collection": name})

            print(f"- {name}")
            print(f"  points: {points}")
            print(f"  vectors: {vectors}")
            print(f"  indexed_vectors: {indexed}")
            print(f"  segments: {segments}")
            print(f"  status: {status}")
    except Exception as exc:
        print(f"\ncollections: error: {exc}")


if __name__ == "__main__":
    main()
