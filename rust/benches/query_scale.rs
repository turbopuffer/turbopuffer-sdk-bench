use criterion::{Criterion, criterion_group, criterion_main};
use futures::StreamExt;
use futures::stream;
use turbopuffer::Namespace;
use turbopuffer::types::{RankBy, Row};

use turbopuffer_sdk_bench_rust as util;

const NUM_DOCS_PER_NAMESPACE: usize = 256;
const NUM_NAMESPACES: usize = 64;
const NUM_QUERIES: usize = 4096;
const NUM_WORKERS: usize = 128;

async fn do_upsert(ns: &Namespace, i: usize, docs: Vec<Row>) {
    println!(
        "[{}/{NUM_NAMESPACES}] Upserting {} documents into namespace {}",
        i + 1,
        docs.len(),
        ns.id()
    );
    util::upsert_into(ns, docs).await;
}

async fn do_query(ns: &Namespace, i: usize) {
    println!("[{}/{NUM_QUERIES}] Querying namespace {}", i + 1, ns.id());
    let results = ns
        .query()
        .rank_by(RankBy::ann("vector", util::random_vector()))
        .top_k(1)
        .send()
        .await
        .expect("query failed");
    assert_eq!(
        results.rows.len(),
        1,
        "expected exactly one result, got {}",
        results.rows.len()
    );
}

async fn run_query_scale_benchmark(namespaces: &[Namespace]) {
    stream::iter(0..NUM_QUERIES)
        .for_each_concurrent(NUM_WORKERS, |i| {
            do_query(&namespaces[i % NUM_NAMESPACES], i)
        })
        .await;
}

fn bench_query_scale(c: &mut Criterion) {
    let rt = util::runtime();
    let client = util::new_client();
    let namespaces: Vec<_> = (0..NUM_NAMESPACES)
        .map(|_| util::random_namespace(&client))
        .collect();

    // Generate a small number of documents in a fair number of namespaces
    // outside the benchmark function.
    let docs = util::random_documents(NUM_DOCS_PER_NAMESPACE, 8);
    rt.block_on(
        stream::iter(namespaces.iter().enumerate())
            .for_each_concurrent(NUM_WORKERS, |(i, ns)| do_upsert(ns, i, docs.clone())),
    );

    c.bench_function("query_scale", |b| {
        b.to_async(&rt)
            .iter(|| run_query_scale_benchmark(&namespaces))
    });
}

criterion_group! {
    name = benches;
    config = util::criterion();
    targets = bench_query_scale
}
criterion_main!(benches);
