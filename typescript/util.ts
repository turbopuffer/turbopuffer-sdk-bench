import { randomBytes } from 'crypto';
import { v4 as uuidv4 } from 'uuid';
import { Turbopuffer, Namespace, UpsertRows, AttributeType } from '@turbopuffer/turbopuffer';

const VECTOR_DIMS = 1536;

export function newClient(): Turbopuffer {
    return new Turbopuffer({
        apiKey: process.env.TURBOPUFFER_API_KEY!
    });
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

export function randomDocument(textContentSize: number): object {
    return {
        id: uuidv4(),
        vector: randomVector(),
        content: randomString(textContentSize),
    };
}

export function randomDocuments(numDocs: number, textContentSize: number): UpsertRows {
    return Array.from({ length: numDocs }, () => randomDocument(textContentSize)) as UpsertRows;
}

export function randomNamespace(client: Turbopuffer): any {
    return client.namespace(`turbopuffer-sdk-bench-typescript-${randomString(12)}`);
}

export async function upsertInto(ns: Namespace, rows: UpsertRows) {
    await ns.write({
        upsert_rows: rows,
        distance_metric: 'cosine_distance'
    });
}
