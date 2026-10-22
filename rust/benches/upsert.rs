use criterion::{BatchSize, Criterion, criterion_group, criterion_main};

use turbopuffer_sdk_bench_rust as util;

fn bench_upsert(c: &mut Criterion) {
    let rt = util::runtime();

    // Create and reuse single client instance to avoid including init / TLS
    // handshake in benchmark.
    let client = util::new_client();

    // Generate documents outside the benchmark function.
    let upsert_docs = util::random_documents(1024, 8);

    c.bench_function("upsert", |b| {
        b.to_async(&rt).iter_batched(
            // The SDK takes ownership of rows, so clone them outside the
            // measured routine.
            || upsert_docs.clone(),
            |docs| {
                let ns = util::random_namespace(&client);
                async move { util::upsert_into(&ns, docs).await }
            },
            BatchSize::PerIteration,
        )
    });
}

criterion_group! {
    name = benches;
    config = util::criterion();
    targets = bench_upsert
}
criterion_main!(benches);
