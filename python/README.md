# turbopuffer SDK Benchmarks / Python

Get a baseline:

```
rm -f bench-baseline.json
uv run bench_upsert.py -o bench-baseline.json
uv run bench_query.py --append bench-baseline.json
uv run bench_query_scale.py --append bench-baseline.json
```

Run an experiment:

```
rm -f bench-experiment.json
uv run bench_upsert.py -o bench-experiment.json
uv run bench_query.py --append bench-experiment.json
uv run bench_query_scale.py --append bench-experiment.json
```

Compare the results:

```
uv run pyperf compare_to --table bench-baseline.json bench-experiment.json
```
