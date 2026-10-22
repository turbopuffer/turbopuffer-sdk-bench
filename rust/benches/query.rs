use criterion::{Criterion, criterion_group, criterion_main};
use turbopuffer::Namespace;
use turbopuffer::types::RankBy;

use turbopuffer_sdk_bench_rust as util;

const NUM_DOCS: usize = 256;

async fn run_query_benchmark(ns: &Namespace) {
    let results = ns
        .query()
        .rank_by(RankBy::ann("vector", util::random_vector()))
        .include_attributes(true)
        .top_k(1200)
        .send()
        .await
        .expect("query failed");
    assert_eq!(
        results.rows.len(),
        NUM_DOCS,
        "expected {NUM_DOCS} results, got {}",
        results.rows.len()
    );
}

fn bench_query(c: &mut Criterion) {
    let rt = util::runtime();
    let query_ns = util::random_namespace(&util::new_client());

    // Generate some document to query outside the benchmark function.
    rt.block_on(util::upsert_into(
        &query_ns,
        util::random_documents(NUM_DOCS, 8096),
    ));

    c.bench_function("query", |b| {
        b.to_async(&rt).iter(|| run_query_benchmark(&query_ns))
    });
}

criterion_group! {
    name = benches;
    config = util::criterion();
    targets = bench_query
}
criterion_main!(benches);
