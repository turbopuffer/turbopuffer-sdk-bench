import { bench, beforeAll } from 'vitest';
import { newClient, randomDocuments, randomNamespace, randomVector, upsertInto } from './util.js';

const NUM_DOCS = 256;
let queryNs: any;

beforeAll(async () => {
    const client = newClient();
    queryNs = randomNamespace(client);
    const docs = randomDocuments(NUM_DOCS, 8096);
    await upsertInto(queryNs, docs);
});

bench('query', async () => {
    const results = await queryNs.query({
        rank_by: ["vector", "ANN", randomVector()],
        include_attributes: true,
        top_k: 1200
    });

    if (results.rows.length !== NUM_DOCS) {
        throw new Error(`expected ${NUM_DOCS} results, got ${results.length}`);
    }
}, {
    iterations: 10,
    warmupIterations: 1
});
