import Turbopuffer from '@turbopuffer/turbopuffer'
import { Command } from 'commander'
import { fork, ChildProcess } from 'child_process'
import { BatchEvent } from './worker';

// TODO:
// - collect avg number of retries per batch
// - collect avg estimated_seconds_remaining
// over time too ^

const DEFAULT_CONCURRENCY: number = 16;
const DEFAULT_NUM_DOCS: number = 1024 * 1024 * 10;
const DEFAULT_DOCS_PER_BATCH: number = 20480;

const DEFAULT_VEC_SIZE: number = 1024;
const DEFAULT_CONTENT_SIZE: number = 64;
const DEFAULT_RETRY_AFTER: number = 10;
const DEFAULT_GRANULARITY: number = 10;

type Conf = {
	numDocs: number,
	docsPerBatch: number,
	vecSize: number,
	contentSize: number,
	concurrency: number,
	retryAfter: number,
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
	docs: number;
	misses: number;
	missPerSec: number;
	docPerSec: number;
};

enum EventKind {
	WriteEvent,
	MissEvent,
};

type Event = {
	kind: EventKind;
	docs: number;
};

const cmd = new Command();
cmd
	.option('--name <n>', 'A unique name for this benchmark', 'benchmark')
	.option('--num-docs <number>', 'Number of docs to ingest', DEFAULT_NUM_DOCS.toString())
	.option('--docs-per-batch <number>', 'Number of docs in a batch', DEFAULT_DOCS_PER_BATCH.toString())
	.option('--vec-size <number>', 'Number of dimensions for vectors', DEFAULT_VEC_SIZE.toString())
	.option('--content-size <number>', 'Number of bytes in content attr', DEFAULT_CONTENT_SIZE.toString())
	.option('--retry-after <number>', 'Number of seconds to wait before retrying a 429', DEFAULT_RETRY_AFTER.toString())
	.option('--concurrency <number>', 'Number of concurrent processes to use', DEFAULT_CONCURRENCY.toString())
	.option('--granularity <number>', 'Number of seconds to track stats at', DEFAULT_GRANULARITY.toString())
	.parse();
const opts = cmd.opts();

const conf: Conf = {
	numDocs: parseInt(opts.numDocs),
	docsPerBatch: parseInt(opts.docsPerBatch),
	vecSize: parseInt(opts.vecSize),
	contentSize: parseInt(opts.contentSize),
	retryAfter: parseInt(opts.retryAfter),
	concurrency: parseInt(opts.concurrency),
	gran: parseInt(opts.granularity),
};

console.log(`config: ${JSON.stringify(conf)}`);
const numBatches = conf.numDocs / conf.docsPerBatch;
const batchPerConc = numBatches / conf.concurrency;
const workerPromise = new Promise<void>((resolve, reject) => {
	const worker = fork('worker.ts', [
		'--docs-per-batch', conf.docsPerBatch.toString(),
		'--num-batches', batchPerConc.toString(),
		'--vec-size', conf.vecSize.toString(),
		'--content-size', conf.contentSize.toString(),
		'--retry-after', conf.retryAfter.toString(),
	], {
		execArgv: ['--import', 'tsx/esm']
	});

	worker.on('message', (msg: BatchEvent) => {
		switch (msg) {
			case BatchEvent.Write:
				console.log("write event");
				break;
			case BatchEvent.Miss:
				console.log("miss event");
				break;
			default:
				reject("invalid batch event");
				break;
		}
		resolve();
	});

	worker.on('error', (err) => {
		console.log("got an error");
		console.log(err);
		reject(err);
	});

	worker.on('exit', (code) => {
		console.log(`worker exited with ${code}`);
	});
});

await workerPromise;
