package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/turbopuffer/turbopuffer-go"
)

// TODO:
// - maybe templates for benches, similar to the server benchmark
// - go's prof tooling might be useful for checking cpu bound client side stuff like compression
// - options for benchmarks:
//   - sizes of stuff (num namespaces, size of vecs, etc)
//   - compression or not
//   - amount of batching/concurrency
//   - retry mechanism
// - column vs row format
// - figure out reasonable baseline behavior for retrys (sounds like a lot of users just don't do anything?)

// NOTE:
// questions:
// - how many docs to ingest? want a representative workload without it being like, an *actual*
//   giant ingest since we're gonna be benching in prod?
// metrics to collect:
// - 429 per sec
// - docs per sec

const TOO_MANY_REQS int = 429

type Conf struct {
	DocsPerNS    int
	DocsPerBatch int
	NumNS        int
	VecSize      int
	ContentSize  int
	RetryAfter   time.Duration
}

// this is just writing a bunch of docs one at a time to a single namespace with no concurrency
// i don't imagine this would actually exist in the wild, but might be a good worst case baseline
func IngestNaive(eChan chan<- Event, conf Conf) {
	ctx := context.Background()
	client := NewClient()
	ns := client.Namespace(RandomNamespace())

	for range conf.DocsPerNS {
		// for now im just going to generate the docs/requests inline since actual
		// clients will be doing some amount of work here when ingesting anyway
		_, err := ns.Write(ctx, turbopuffer.NamespaceWriteParams{
			UpsertRows: []turbopuffer.RowParam{RandomDoc(conf.VecSize, conf.ContentSize)},
		})
		if err != nil {
			var typedErr turbopuffer.Error
			if errors.As(err, typedErr) {
				if typedErr.StatusCode == TOO_MANY_REQS {
					eChan <- Event{
						kind: MissEvent,
						docs: 1,
					}
					// since this is a worst case, we can just not do any kind
					// of retry
				} else {
					panic(err)
				}
			} else {
				panic(err)
			}
		} else {
			eChan <- Event{
				kind: WriteEvent,
				docs: 1,
			}
		}
	}
}

// this will be a more reasonable baseline, basically a handful of namespaces, with a go routine per
// namespace, and a little bit of batching (simulating maybe shipping a batch of records per src
// file or something), and a super simple retry loop
func IngestBaseline(eChan chan<- Event, conf Conf) {
	ctx := context.Background()
	client := NewClient()

	var wg sync.WaitGroup
	for range conf.NumNS {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ns := client.Namespace(RandomNamespace())
			numBatches := conf.DocsPerNS / conf.DocsPerBatch
			for range numBatches {
				batch := RandomBatch(conf.DocsPerBatch, conf.VecSize, conf.ContentSize)
				// super simple retry loop
				for {
					_, err := ns.Write(ctx, turbopuffer.NamespaceWriteParams{
						UpsertRows: batch,
					})
					if err != nil {
						var typed *turbopuffer.Error
						if errors.As(err, &typed) {
							if typed.StatusCode == TOO_MANY_REQS {
								eChan <- Event{
									kind: MissEvent,
									docs: len(batch),
								}
								time.Sleep(conf.RetryAfter)
								continue
							} else {
								panic(err)
							}
						} else {
							panic(err)
						}
					} else {
						eChan <- Event{
							kind: WriteEvent,
							docs: len(batch),
						}
					}
				}
			}
		}()
	}

	wg.Wait()
}

// this will be iteratively adjusted and end up being whatever gets put in the docs/guide
func IngestGuide() {

}

// the idea here is to have something that can be used safely from multiple goroutines, and you
// can just basically chuck records at it and it will manage batching, parallel requests, retrys, etc
// TODO: error propagating, also i think the client lib handles retrys internally to a certain extent

type Ingest struct {
	client     turbopuffer.Client
	namespaces map[string]*IngestNamespace
	nsLock     sync.RWMutex
	opts       IngestOpts
}

func (i *Ingest) Push(row turbopuffer.Row, namespace string) {
	// first check if we're already managing the namespace
	i.nsLock.RLock()
	if ns, exists := i.namespaces[namespace]; exists {
		ns.lock.Lock()
		defer ns.lock.Unlock()
		defer i.nsLock.RUnlock()

		ns.currentBatch = append(ns.currentBatch, row)
		if len(ns.currentBatch) >= i.opts.BatchSize {
			// batch is full, send the request
			go writeBatch(&ns.ns, ns.currentBatch)
			ns.currentBatch = make([]turbopuffer.Row, 0, i.opts.BatchSize)
		}
	} else {
		i.nsLock.RUnlock()
		i.nsLock.Lock()
		defer i.nsLock.Unlock()

		batch := make([]turbopuffer.Row, 0, i.opts.BatchSize)
		batch = append(batch, row)
		i.namespaces[namespace] = &IngestNamespace{
			ns:           i.client.Namespace(namespace),
			currentBatch: batch,
		}
	}

}
func (i *Ingest) Flush() {}

type IngestOpts struct {
	BatchSize int
}
type IngestNamespace struct {
	ns           turbopuffer.Namespace
	lock         sync.Mutex
	currentBatch []turbopuffer.Row
}

func writeBatch(ns *turbopuffer.Namespace, batch []turbopuffer.Row) {
}
