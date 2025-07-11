import Turbopuffer from '@turbopuffer/turbopuffer'
import { Row } from '@turbopuffer/turbopuffer/resources/namespaces';
import { Command } from 'commander'

type WorkerConf = {
	docsPerBatch: number,
	numBatches: number,
	vecSize: number,
	contentSize: number,
	retryAfter: number,
};

const run = async () => {
	const cmd = new Command();
	cmd
		.option('--docs-per-batch <number>')
		.option('--num-batches <number>')
		.option('--vec-size <number>')
		.option('--content-size <number>')
		.option('--retry-after <number>')
		.parse();
	const opts = cmd.opts();

	const conf: WorkerConf = {
		docsPerBatch: parseInt(opts.docsPerBatch),
		numBatches: parseInt(opts.numBatches),
		vecSize: parseInt(opts.vecSize),
		contentSize: parseInt(opts.contentSize),
		retryAfter: parseInt(opts.retryAfter),
	};
	const client = new Turbopuffer({
		apiKey: process.env['TURBOPUFFER_API_KEY'],
		region: 'aws-test-1',
		maxRetries: 0,
	});
	const ns = client.namespace('andrew-bench');

	for (let i = 0; i < conf.numBatches; i++) {
		const batch = randomBatch(conf.docsPerBatch, conf.vecSize, conf.contentSize);
		let t = 0;
		while (true) {
			t += 1;
			ns.write({
				upsert_rows: batch,
				distance_metric: 'cosine_distance',
			}, {
				maxRetries: 0,
			}).catch(async (err) => {
				if (err instanceof Turbopuffer.APIError) {
					if (err.status == 429) {
						console.log(`429 on try ${t}`);
						process.send!(BatchEvent.Miss);
					} else {
						console.log(`non 429 err on try ${t}`);
					}

					await new Promise(
						resolve => setTimeout(
							resolve,
							conf.retryAfter * 1000,
						),
					);
				} else {
					console.error(`non typed err: ${err}`);
					process.exit(1);
				}
			}).then(_ => { });
		}
	}
}

export enum BatchEvent {
	Write,
	Miss,
};


const randomBatch = (docsPerBatch: number, vecSize: number, contentSize: number): Row[] => {
	let batch = [];
	for (let i = 0; i < docsPerBatch; i++) {

	}
}

await run();
