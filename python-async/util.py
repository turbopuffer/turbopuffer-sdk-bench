import asyncio
import random
import string
from threading import Thread
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
        "content": random_string(text_content_size),
    }

def random_documents(num_docs, text_content_size):
    return [random_document(text_content_size) for _ in range(num_docs)]

def new_client():
    return tpuf.AsyncTurbopuffer(region="gcp-us-central1")

def random_namespace(client):
    return client.namespace(f"turbopuffer-sdk-bench-python-{random_string(12)}")

async def upsert_into(ns, docs):
    await ns.write(
        upsert_rows=docs,
        distance_metric='cosine_distance'
    )

def start_async_thread():
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    Thread(target=loop.run_forever, daemon=True).start()


def run_on_async_thread(coro):
    loop = asyncio.get_event_loop()
    future = asyncio.run_coroutine_threadsafe(coro, loop)
    return future.result()


def wrap_async_thread(fn):
    def wrapper(*args, **kwargs):
        return run_on_async_thread(fn(*args, **kwargs))
    return wrapper
