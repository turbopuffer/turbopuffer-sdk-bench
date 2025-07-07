package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// TODO: this is probably way too low, get a decent number for this
const DEFAULT_DOCS_PER_NS int = 1024
const DEFAULT_NUM_NS int = 16
const DEFAULT_DOCS_PER_BATCH int = 32
const DEFAULT_GRANULARITY int = 1000

var workload = flag.String("workload", "", "which workload to benchmark")
var gran = flag.Int("gran", DEFAULT_GRANULARITY, "granularity to collect stats at in milleseconds")
var docsPerNS = flag.Int("docs-per-ns", DEFAULT_DOCS_PER_NS, "number of docs to ingest per namespace")
var numNS = flag.Int("num-ns", DEFAULT_NUM_NS, "number of namespaces")
var docsPerBatch = flag.Int("docs-per-batch", DEFAULT_DOCS_PER_BATCH, "number of docs in a batch")
var out = flag.String("o", "bench.json", "output file to dump stats to")

func main() {
	flag.Parse()

	eChan := make(chan Event)
	snapChan := make(chan []Snap, 1)
	doneChan := make(chan struct{}, 1)
	granDur := time.Duration(*gran) * time.Millisecond
	go CollectSnaps(eChan, granDur, doneChan, snapChan)

	start := time.Now()
	switch *workload {
	case "naive":
		IngestNaive(*docsPerNS, eChan)
	case "baseline":
		IngestBaseline(*numNS, *docsPerNS, *docsPerBatch, eChan)
	default:
		fmt.Printf("invalid workload: %s", *workload)
		return
	}
	totalDur := time.Since(start)

	doneChan <- struct{}{}
	snaps := <-snapChan
	totalDocs := 0
	totalMisses := 0
	for _, snap := range snaps {
		totalDocs += snap.Docs
		totalMisses += snap.Misses
	}

	bench := Bench{
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

	err = os.WriteFile(*out, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

type Bench struct {
	Snaps       []Snap        `json:"snaps"`
	TotalDur    time.Duration `json:"total_duration"`
	TotalDocs   int           `json:"total_docs"`
	TotalMisses int           `json:"total_misses"`
	MissPerSec  float64       `json:"miss_per_sec"`
	DocPerSec   float64       `json:"doc_per_sec"`
}
