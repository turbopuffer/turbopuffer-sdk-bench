import { v4 as uuidv4 } from 'uuid';
import { Turbopuffer } from '@turbopuffer/turbopuffer';
import { Row } from '@turbopuffer/turbopuffer/resources/namespaces';

const VECTOR_DIMS = 1536;

export function newClient(): Turbopuffer {
    return new Turbopuffer();
}

export function randomFloat(): number {
    return Math.random();
}

export function randomString(size: number): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    return Array.from({ length: size }, () =>
        chars.charAt(Math.floor(Math.random() * chars.length))
    ).join('');
}

export function randomVector(): number[] {
    return Array.from({ length: VECTOR_DIMS }, () => randomFloat());
}

export function randomDocument(textContentSize: number): Row {
    return {
        id: uuidv4(),
        vector: randomVector(),
        content: randomString(textContentSize),
    };
}

export function randomDocuments(numDocs: number, textContentSize: number): Row[] {
    return Array.from({ length: numDocs }, () => randomDocument(textContentSize));
}

export function randomNamespace(client: Turbopuffer): any {
    return client.namespace(`turbopuffer-sdk-bench-typescript-${randomString(12)}`);
}

export async function upsertInto(ns: Turbopuffer.Namespace, rows: Row[]) {
    await ns.write({
        upsert_rows: rows,
        distance_metric: 'cosine_distance'
    });
}
