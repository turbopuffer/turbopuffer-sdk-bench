import { bench, beforeAll } from 'vitest';
import { newClient, randomDocuments, randomNamespaceId, randomVector, upsertInto } from './util.js';
import { Namespace } from '@turbopuffer/turbopuffer';
import { limitFunction } from 'p-limit';

const NUM_DOCS_PER_NAMESPACE = 256;
const NUM_NAMESPACES = 64;
const NUM_QUERIES = 4096;
const NUM_WORKERS = 128;

let namespaceIds: string[];
let namespaces: Namespace[];

beforeAll(async () => {
    const client = newClient();
    namespaceIds = Array.from({ length: NUM_NAMESPACES }, () => randomNamespaceId());
    namespaces = namespaceIds.map((id) => client.namespace(id));
    const docs = randomDocuments(NUM_DOCS_PER_NAMESPACE, 8);

    // Upsert documents into all namespaces concurrently.
    const upsertPromises = namespaces.map(async (ns, i) => {
        console.log(`[${i + 1}/${NUM_NAMESPACES}] Upserting ${docs.length} documents into namespace ${namespaceIds[i]}`);
        return upsertInto(ns, docs);
    });

    await Promise.all(upsertPromises);
}, 60_000);

async function doQuery(ns: any, id: string, i: number): Promise<void> {
    console.log(`[${i + 1}/${NUM_QUERIES}] Querying namespace ${id}`);
    const results = await ns.query({
        rank_by: ["vector", "ANN", randomVector()],
        top_k: 1
    });
    if (results.rows.length !== 1) {
        throw new Error(`expected exactly one result, got ${results.rows.length}`);
    }
}

bench('query_scale', async () => {
    const doQueryLimited = limitFunction(doQuery, { concurrency: NUM_WORKERS });
    const queries = Array.from(
        { length: NUM_QUERIES },
        (_, i) => doQueryLimited(namespaces[i % NUM_NAMESPACES], namespaceIds[i % NUM_NAMESPACES], i)
    );
    await Promise.all(queries);
}, {
    iterations: 3,
    warmupIterations: 1
});
