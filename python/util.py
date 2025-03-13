import random
import string
import uuid

import turbopuffer as tpuf

VECTOR_DIMS = 1536

def random_float():
    return random.random()

def random_string(size):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=size))

def random_vector():
    return [random_float() for _ in range(VECTOR_DIMS)]

def random_document(text_content_size):
    return {
        "id": str(uuid.uuid4()),
        "vector": random_vector(),
        "attributes": {
            "content": random_string(text_content_size)
        }
    }

def random_documents(num_docs, text_content_size):
    return [random_document(text_content_size) for _ in range(num_docs)]

def random_namespace():
    return tpuf.Namespace(f"turbopuffer-sdk-bench-python-{random_string(12)}")

def upsert_into(ns, docs):
    ns.upsert(
        docs,
        distance_metric='cosine_distance'
    )
