use std::time::Duration;

use criterion::Criterion;
use rand::Rng;
use rand::distr::Alphanumeric;
use turbopuffer::types::{AttributeSchemaConfig, DistanceMetric, Row};
use turbopuffer::{Client, Namespace};
use uuid::Uuid;

pub const VECTOR_DIMS: usize = 1536;

/// Criterion configured to run a single warmup iteration followed by ten
/// samples of one iteration each, like the other languages' benchmarks.
pub fn criterion() -> Criterion {
    Criterion::default()
        .sample_size(10)
        .warm_up_time(Duration::from_millis(1))
        .measurement_time(Duration::from_millis(1))
}

/// A single-threaded runtime, since benchmarks are single-threaded.
pub fn runtime() -> tokio::runtime::Runtime {
    tokio::runtime::Builder::new_current_thread()
        .enable_all()
        .build()
        .expect("failed to build Tokio runtime")
}

pub fn random_float() -> f32 {
    rand::random()
}

pub fn random_string(size: usize) -> String {
    rand::rng()
        .sample_iter(&Alphanumeric)
        .take(size)
        .map(char::from)
        .collect()
}

pub fn random_vector() -> Vec<f32> {
    (0..VECTOR_DIMS).map(|_| random_float()).collect()
}

pub fn random_document(text_content_size: usize) -> Row {
    Row::new(Uuid::new_v4().to_string())
        .vector(random_vector())
        .insert("content", random_string(text_content_size))
}

pub fn random_documents(num_docs: usize, text_content_size: usize) -> Vec<Row> {
    (0..num_docs)
        .map(|_| random_document(text_content_size))
        .collect()
}

pub fn new_client() -> Client {
    Client::from_env().expect("failed to build client")
}

pub fn random_namespace(client: &Client) -> Namespace {
    client.namespace(format!("turbopuffer-sdk-bench-rust-{}", random_string(12)))
}

pub async fn upsert_into(ns: &Namespace, docs: Vec<Row>) {
    ns.write()
        .upsert_rows(docs)
        .distance_metric(DistanceMetric::CosineDistance)
        // Content isn't filtered on, and is too large to be filterable in the
        // query benchmark.
        .schema([(
            "content",
            AttributeSchemaConfig::new("string").filterable(false),
        )])
        .send()
        .await
        .expect("upsert failed");
}
