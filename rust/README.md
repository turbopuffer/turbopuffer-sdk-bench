# turbopuffer SDK Benchmarks / Rust

Get a baseline:

```
cargo bench --bench upsert -- --save-baseline baseline
cargo bench --bench query -- --save-baseline baseline
cargo bench --bench query_scale -- --save-baseline baseline
```

Run an experiment and compare the results:

```
cargo bench --bench upsert -- --baseline baseline
cargo bench --bench query -- --baseline baseline
cargo bench --bench query_scale -- --baseline baseline
```

To benchmark a local checkout of the SDK, add a patch to `Cargo.toml`:

```toml
[patch.crates-io]
turbopuffer = { path = "../../turbopuffer-rust" }
```
