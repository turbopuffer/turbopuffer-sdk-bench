package main

import (
	"math/rand/v2"

	"github.com/turbopuffer/turbopuffer-go"
)

func NewClient() turbopuffer.Client {
	return turbopuffer.NewClient()
}

func RandomNamespace() string {
	return RandomString(12)
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
