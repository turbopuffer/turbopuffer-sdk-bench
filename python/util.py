import asyncio
import random
import string
from threading import Thread
import uuid

from aiohttp import ClientSession
import httpx
from httpx_aiohttp import AiohttpTransport
from turbopuffer import AsyncTurbopuffer

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
    return f"turbopuffer-sdk-bench-python-{random_string(12)}"

async def upsert_into(tpuf, ns, docs):
    await tpuf.namespaces.upsert(
        namespace=ns,
        documents={
            "upserts": docs,
            "distance_metric": "cosine_distance",
        },
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


async def make_client():
    return AsyncTurbopuffer(
        http_client=httpx.AsyncClient(
            transport=AiohttpTransport(client=ClientSession()),
        )
    )
