import requests

def get_vector_count(collection_name):
    # Query the Prometheus instance we just set up
    resp = requests.get("http://localhost:9097/api/v1/query", 
                        params={'query': f'qdrant_collection_vectors_total{{collection="{collection_name}"}}'})

    data = resp.json()
    results = data.get('data', {}).get('result', [])
    if not results:
        return 0 # Return 0 if collection not found or no metric matches
    return results[0]['value'][1]

print(f"Current usage: {get_vector_count('biomedical_papers')} vectors")