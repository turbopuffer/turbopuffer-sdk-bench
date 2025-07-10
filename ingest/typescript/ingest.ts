import Turbopuffer from '@turbopuffer/turbopuffer'
import { isMainThread, Worker } from 'node:worker_threads'

type Conf = {
	numDocs: number,
	docsPerBatch: number,
	vecSize: number,
	contentSize: number,
	concurrency: number,
	retryAfter: number,
	maxRetries: number,
	gran: number,
};

type Bench = {
	conf: Conf,
	snaps: Snap[],
	totalDur: number,
	totalDocs: number,
	totalMisses: number,
	missPerSec: number,
	docPerSec: number,
};

type Snap = {

};

const runBench = (conf: Conf) => {
	const client = new Turbopuffer({});
	const ns = client.namespace("andrew-bench");

	const numBatches = conf.numDocs / conf.docsPerBatch;
	const batchPerConc = numBatches / conf.concurrency;
}

if (isMainThread) {

}
