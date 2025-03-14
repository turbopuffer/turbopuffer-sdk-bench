import sys

import pyperf

import util


@util.wrap_async_thread
async def run_upsert_benchmark(tpuf, docs):
    await util.upsert_into(tpuf, util.random_namespace(), docs)


def main():
    # Convince pyperf to keep the `TURBOPUFFER_API_KEY` environment variable
    # through benchmark subprocesses.
    sys.argv += ["--copy-env"]

    runner = pyperf.Runner(processes=1, warmups=1, values=10)
    runner.parse_args()

    util.start_async_thread()
    tpuf = util.run_on_async_thread(util.make_client())
    upsert_docs = []

    # Generate documents outside the benchmark function.
    if runner.args.worker:
        upsert_docs = util.random_documents(num_docs=1024, text_content_size=8)

    runner.bench_func(
        "upsert",
        run_upsert_benchmark,
        tpuf,
        upsert_docs,
    )

if __name__ == "__main__":
    main()
