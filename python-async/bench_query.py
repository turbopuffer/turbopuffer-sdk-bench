import sys

import pyperf

import util

NUM_DOCS = 256

@util.wrap_async_thread
async def run_query_benchmark(ns):
    results = await ns.query(
        rank_by=["vector", "ANN", util.random_vector()],
        include_attributes=True,
        top_k=1200,
    )
    assert len(results.rows) == NUM_DOCS, f"expected {NUM_DOCS} results, got {len(results)}"


def main():
    # Convince pyperf to keep the `TURBOPUFFER_API_KEY` environment variable
    # through benchmark subprocesses.
    sys.argv += ["--copy-env"]

    runner = pyperf.Runner(processes=1, warmups=1, values=10)
    runner.parse_args()

    util.start_async_thread()
    query_ns = util.random_namespace(util.new_client())

    # Generate some document to query outside the benchmark function.
    if runner.args.worker:
        util.run_on_async_thread(util.upsert_into(query_ns, util.random_documents(num_docs=NUM_DOCS, text_content_size=8096)))

    runner.bench_func(
        "query",
        run_query_benchmark,
        query_ns,
    )


if __name__ == "__main__":
    main()
