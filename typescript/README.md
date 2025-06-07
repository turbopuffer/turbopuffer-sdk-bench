# turbopuffer SDK Benchmarks / TypeScript

Get a baseline:

```
rm -f bench-baseline.json
npx vitest bench upsert.bench.ts --outputJson=bench-upsert-baseline.json
npx vitest bench query.bench.ts --outputJson=bench-query-baseline.json
npx vitest bench query_scale.bench.ts --outputJson=bench-query-scale-baseline.json
```

Run an experiment and compare the results:

```
rm -f bench-experiment.json
npx vitest bench upsert.bench.ts --compare=bench-upsert-baseline.json
npx vitest bench query.bench.ts --compare=bench-query-baseline.json
npx vitest bench query_scale.bench.ts --compare=bench-query-scale-baseline.json
```
