import { bench } from 'vitest';
import { newClient, randomDocuments, randomNamespace, upsertInto } from './util.js';

const client = newClient();
const docs = randomDocuments(1024, 8);

bench('upsert', async () => {
    const ns = randomNamespace(client);
    await upsertInto(ns, docs);
}, {
    iterations: 10,
    warmupIterations: 1
});
