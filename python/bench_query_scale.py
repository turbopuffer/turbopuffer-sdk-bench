import asyncio
import os
import sys
import traceback

import pyperf

import util

NUM_DOCS_PER_NAMESPACE = 256
NUM_NAMESPACES = 64
NUM_QUERIES = 4096
NUM_WORKERS = 128

semaphore = asyncio.Semaphore(NUM_WORKERS)

def fail_hard(fn):
    async def wrapper(*args, **kwargs):
        try:
            return await fn(*args, **kwargs)
        except Exception as e:
            traceback.print_exc()
            os._exit(1)
    return wrapper


@fail_hard
async def do_upsert(tpuf, ns, i, docs):
    async with semaphore:
        print(f"[{i+1}/{NUM_NAMESPACES}] Upserting {len(docs)} documents into namespace {ns}")
        await util.upsert_into(tpuf, ns, docs)


@fail_hard
async def do_query(tpuf, ns, i):
    async with semaphore:
        print(f"[{i+1}/{NUM_QUERIES}] Querying namespace {ns}")
        results = await tpuf.namespaces.query(
            namespace=ns,
            vector=util.random_vector(),
            top_k=1,
        )
        assert len(results) == 1, f"expected exactly one result, got {len(results)}"


@util.wrap_async_thread
async def run_query_scale_benchmark(tpuf, namespaces):
    async with asyncio.TaskGroup() as tg:
        for i in range(NUM_QUERIES):
            tg.create_task(do_query(tpuf, namespaces[i % NUM_NAMESPACES], i))


def main():
    # Convince pyperf to keep the `TURBOPUFFER_API_KEY` environment variable
    # through benchmark subprocesses.
    sys.argv += ["--copy-env"]

    runner = pyperf.Runner(processes=1, warmups=0, values=10)
    runner.parse_args()

    util.start_async_thread()
    tpuf = util.run_on_async_thread(util.make_client())
    namespaces = [util.random_namespace() for _ in range(NUM_NAMESPACES)]

    # Generate a small number of documents in a fair number of namespaces
    # outside the benchmark function.
    if runner.args.worker:
        docs = util.random_documents(num_docs=NUM_DOCS_PER_NAMESPACE, text_content_size=8)
        async def submit():
            async with asyncio.TaskGroup() as tg:
                for i, ns in enumerate(namespaces):
                    tg.create_task(do_upsert(tpuf, ns, i, docs))
        util.run_on_async_thread(submit())

    runner.bench_func(
        "query_scale",
        run_query_scale_benchmark,
        tpuf,
        namespaces,
    )


if __name__ == "__main__":
    main()
