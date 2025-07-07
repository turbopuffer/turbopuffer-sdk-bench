package main

import "time"

type Snap struct {
	Docs       int
	Misses     int
	MissPerSec float64
	DocPerSec  float64
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

	secs := gran.Seconds()
	ticker := time.NewTicker(gran)
	for {
		select {
		case <-done:
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
