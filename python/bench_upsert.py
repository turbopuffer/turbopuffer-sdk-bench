import sys

import pyperf

import util

NUM_DOCS = 1024
BASE64_VECTORS = True

def run_upsert_benchmark(docs):
    util.upsert_into(util.random_namespace(), docs)


def main():
    # Convince pyperf to keep the `TURBOPUFFER_API_KEY` environment variable
    # through benchmark subprocesses.
    sys.argv += ["--copy-env"]

    runner = pyperf.Runner(processes=1, warmups=1, values=10)
    runner.parse_args()

    upsert_docs = []

    # Generate documents outside the benchmark function.
    if runner.args.worker:
        upsert_docs = util.random_documents(num_docs=NUM_DOCS, text_content_size=8, base64_vectors=BASE64_VECTORS)

    runner.bench_func(
        "upsert",
        run_upsert_benchmark,
        upsert_docs,
    )


if __name__ == "__main__":
    main()
