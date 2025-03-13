import sys

import pyperf

import util

NUM_DOCS = 256

def run_query_benchmark(ns):
    results = ns.query(
        vector=util.random_vector(),
        include_attributes=True,
        include_vectors=True,
        top_k=1200,
    )
    assert len(results) == NUM_DOCS, f"expected {NUM_DOCS} results, got {len(results)}"


def main():
    # Convince pyperf to keep the `TURBOPUFFER_API_KEY` environment variable
    # through benchmark subprocesses.
    sys.argv += ["--copy-env"]

    runner = pyperf.Runner(processes=1, warmups=1, values=10)
    runner.parse_args()

    query_ns = util.random_namespace()

    # Generate some document to query outside the benchmark function.
    if runner.args.worker:
        util.upsert_into(query_ns, util.random_documents(num_docs=NUM_DOCS, text_content_size=8096))

    runner.bench_func(
        "query",
        run_query_benchmark,
        query_ns,
    )


if __name__ == "__main__":
    main()
