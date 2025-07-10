package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	// "strconv"
	"sync"

	"os"
	"time"

	"github.com/turbopuffer/turbopuffer-go"
	"github.com/turbopuffer/turbopuffer-go/option"
)

const DEFAULT_CONCURRENCY int = 16
const DEFAULT_NUM_DOCS int = 1024 * 1024 * 10
const DEFAULT_DOCS_PER_BATCH int = 20480

const DEFAULT_VEC_SIZE int = 1024
const DEFAULT_CONTENT_SIZE int = 64
const DEFAULT_RETRY_AFTER int = 10
const DEFAULT_MAX_RETRIES int = 2
const DEFAULT_GRANULARITY int = 10

var name = flag.String("name", "", "a unique name for this benchmark")
var numDocs = flag.Int("num-docs", DEFAULT_NUM_DOCS, "number of docs to ingest")
var docsPerBatch = flag.Int("docs-per-batch", DEFAULT_DOCS_PER_BATCH, "number of docs in a batch")
var vecSize = flag.Int("vec-size", DEFAULT_VEC_SIZE, "number of dimensions for vectors")
var contentSize = flag.Int("content-size", DEFAULT_CONTENT_SIZE, "number of bytes in content attr")
var retryAfter = flag.Int("retry-after", DEFAULT_RETRY_AFTER, "number of seconds to wait befor retrying a 429")
var maxRetries = flag.Int("max-retries", DEFAULT_MAX_RETRIES, "max number of retries before giving up on a request")
var concurrency = flag.Int("concurrency", DEFAULT_CONCURRENCY, "number of goroutines to use")
var gran = flag.Int("gran", DEFAULT_GRANULARITY, "number of seconds to track stats at")

func main() {
	flag.Parse()

	eChan := make(chan Event, 1024)
	snapChan := make(chan []Snap)
	doneChan := make(chan struct{})

	granularity := time.Duration(*gran) * time.Second
	conf := Conf{
		NumDocs:      *numDocs,
		DocsPerBatch: *docsPerBatch,
		VecSize:      *vecSize,
		ContentSize:  *contentSize,
		RetryAfter:   time.Duration(*retryAfter) * time.Second,
		MaxRetries:   *maxRetries,
		Concurrency:  *concurrency,
		Gran:         granularity,
	}
	log.Printf("conf:\n%v\n", conf)

	go CollectSnaps(eChan, granularity, doneChan, snapChan)

	log.Printf("starting bench\n")
	start := time.Now()
	runBench(conf, eChan)
	totalDur := time.Since(start)
	log.Printf("bench finished in %v\n", totalDur)

	close(eChan)
	doneChan <- struct{}{}
	snaps := <-snapChan
	totalDocs := 0
	totalMisses := 0
	for _, snap := range snaps {
		totalDocs += snap.Docs
		totalMisses += snap.Misses
	}

	bench := Bench{
		Conf:        conf,
		Snaps:       snaps,
		TotalDur:    totalDur,
		TotalDocs:   totalDocs,
		TotalMisses: totalMisses,
		MissPerSec:  float64(totalMisses) / totalDur.Seconds(),
		DocPerSec:   float64(totalDocs) / totalDur.Seconds(),
	}

	jsonData, err := json.MarshalIndent(bench, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile(fmt.Sprintf("%s.json", *name), jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

const TOO_MANY_REQS int = 429

type Conf struct {
	NumDocs      int           `json:"num_docs"`
	DocsPerBatch int           `json:"docs_per_batch"`
	VecSize      int           `json:"vec_size"`
	ContentSize  int           `json:"content_size"`
	Concurrency  int           `json:"concurrency"`
	RetryAfter   time.Duration `json:"retry_after"`
	MaxRetries   int           `json:"max_retries"`
	Gran         time.Duration `json:"granularity"`
}
type Bench struct {
	Conf        Conf          `json:"conf"`
	Snaps       []Snap        `json:"snaps"`
	TotalDur    time.Duration `json:"total_duration"`
	TotalDocs   int           `json:"total_docs"`
	TotalMisses int           `json:"total_misses"`
	MissPerSec  float64       `json:"miss_per_sec"`
	DocPerSec   float64       `json:"doc_per_sec"`
}

func runBench(conf Conf, eChan chan<- Event) {
	log.Printf("simple retry\n")
	ctx := context.Background()
	client := turbopuffer.NewClient(
		option.WithAPIKey(os.Getenv("TURBOPUFFER_API_KEY")),
		option.WithRegion("aws-test-1"),
		option.WithMaxRetries(0),
	)
	ns := client.Namespace("andrew-bench")

	numBatches := conf.NumDocs / conf.DocsPerBatch
	batchPerConc := numBatches / conf.Concurrency
	var wg sync.WaitGroup
	for range conf.Concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range batchPerConc {
				batch := RandomBatch(
					conf.DocsPerBatch,
					conf.VecSize,
					conf.ContentSize,
				)

				// simple retry loop
				i := 0
				for {
					i += 1
					_, err := ns.Write(
						ctx,
						turbopuffer.NamespaceWriteParams{
							UpsertRows:     batch,
							DistanceMetric: turbopuffer.DistanceMetricCosineDistance,
						},
						option.WithMaxRetries(0),
					)
					if err != nil {
						var typed *turbopuffer.Error
						if errors.As(err, &typed) {
							if typed.StatusCode == TOO_MANY_REQS {
								log.Printf("429 on try %d, ", i)
								eChan <- Event{
									kind: MissEvent,
									docs: len(batch),
								}
								// if strs, ok := typed.Response.Header["Retry-After"]; ok {
								// 	seconds, err := strconv.Atoi(strs[0])
								// 	if err != nil {
								// 		log.Fatalf("%v\n", err)
								// 	}
								// 	d := time.Second * time.Duration(seconds/2)
								// 	log.Printf("retrying in %v (from server)\n", d)
								// 	time.Sleep(d)
								// } else {
								log.Printf("retrying in %v (from conf)\n", conf.RetryAfter)
								time.Sleep(conf.RetryAfter)
								// }
							} else {
								log.Printf("non 429 %v\n", err)
								time.Sleep(conf.RetryAfter)
							}
							continue
						} else {
							log.Fatalf("%v\n", err)
						}
					} else {
						log.Printf("wrote batch on try %d\n", i)
						eChan <- Event{
							kind: WriteEvent,
							docs: len(batch),
						}
						break
					}
				}
			}
		}()
	}

	wg.Wait()
}

type Snap struct {
	Docs       int     `json:"docs"`
	Misses     int     `json:"misses"`
	MissPerSec float64 `json:"miss_per_sec"`
	DocPerSec  float64 `json:"doc_per_sec"`
}

type Event struct {
	kind EventKind
	docs int
}
type EventKind int

const (
	WriteEvent EventKind = iota
	MissEvent
)

func CollectSnaps(
	events <-chan Event,
	gran time.Duration,
	done <-chan struct{},
	res chan<- []Snap,
) {
	snaps := make([]Snap, 0)

	ticker := time.NewTicker(gran)
	secs := gran.Seconds()
	for {
		select {
		case <-done:
			totalDocs := 0
			totalMisses := 0
			for event := range events {
				switch event.kind {
				case WriteEvent:
					totalDocs += event.docs
				case MissEvent:
					totalMisses += 1
				}
			}
			snaps = append(snaps, Snap{
				Docs:       totalDocs,
				Misses:     totalMisses,
				DocPerSec:  float64(totalDocs) / secs,
				MissPerSec: float64(totalMisses) / secs,
			})
			res <- snaps
			return
		case <-ticker.C:
			totalDocs := 0
			totalMisses := 0
			for range len(events) {
				event := <-events
				switch event.kind {
				case WriteEvent:
					totalDocs += event.docs
				case MissEvent:
					totalMisses += 1
				}
			}
			snaps = append(snaps, Snap{
				Docs:       totalDocs,
				Misses:     totalMisses,
				DocPerSec:  float64(totalDocs) / secs,
				MissPerSec: float64(totalMisses) / secs,
			})
		}
	}
}

func RandomVec(size int) []float32 {
	out := make([]float32, size)
	for i := range size {
		out[i] = rand.Float32()
	}
	return out
}

func RandomString(size int) string {
	const charset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	out := make([]byte, size)
	for i := range size {
		out[i] = charset[rand.IntN(len(charset))]
	}
	return string(out)
}

func RandomDoc(vecSize int, contentSize int) turbopuffer.RowParam {
	out := make(map[string]any)
	out["id"] = rand.Uint64()
	out["vector"] = RandomVec(vecSize)
	out["content"] = RandomString(contentSize)
	return out
}

func RandomBatch(batchSize int, vecSize int, contentSize int) []turbopuffer.RowParam {
	out := make([]turbopuffer.RowParam, batchSize)
	for i := range batchSize {
		out[i] = RandomDoc(vecSize, contentSize)
	}
	return out
}
