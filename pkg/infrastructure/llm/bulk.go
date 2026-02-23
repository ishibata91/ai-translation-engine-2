package llm

import (
	"context"
	"log/slog"
	"sync"
)

// ExecuteBulkSync executes multiple LLM requests in parallel using a worker pool.
// It accepts a concurrency limit and processes all requests, collecting results
// in input order. Individual request failures are stored as Response.Success=false
// and do not cause the overall function to return an error.
// The function only returns an error if the context is cancelled before completion.
func ExecuteBulkSync(ctx context.Context, client LLMClient, reqs []Request, concurrency int) ([]Response, error) {
	slog.DebugContext(ctx, "ENTER ExecuteBulkSync",
		slog.Int("total", len(reqs)),
		slog.Int("concurrency", concurrency),
	)
	defer slog.DebugContext(ctx, "EXIT ExecuteBulkSync", slog.Int("total", len(reqs)))

	if concurrency <= 0 {
		concurrency = 1
	}

	results := make([]Response, len(reqs))

	if err := runWorkerPool(ctx, client, reqs, concurrency, results); err != nil {
		return nil, err
	}

	return results, nil
}

// --- Private Methods ---

// runWorkerPool launches worker goroutines up to concurrency, distributes jobs via
// a buffered channel, and blocks until all requests complete or ctx is cancelled.
func runWorkerPool(ctx context.Context, client LLMClient, reqs []Request, concurrency int, results []Response) error {
	slog.DebugContext(ctx, "ENTER runWorkerPool", slog.Int("workers", concurrency))

	type job struct {
		index int
		req   Request
	}

	jobCh := make(chan job, len(reqs))
	for i, req := range reqs {
		jobCh <- job{index: i, req: req}
	}
	close(jobCh)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for j := range jobCh {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		default:
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			defer func() { <-sem }()
			results[j.index] = executeOne(ctx, client, j.index, j.req)
		}(j)
	}

	wg.Wait()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	slog.DebugContext(ctx, "EXIT runWorkerPool")
	return nil
}

// executeOne sends a single request to the LLM client and returns a Response.
// On error, it returns a Response with Success=false and Error set, without propagating the error.
func executeOne(ctx context.Context, client LLMClient, index int, req Request) Response {
	slog.DebugContext(ctx, "ENTER executeOne", slog.Int("index", index))

	resp, err := client.Complete(ctx, req)
	if err != nil {
		slog.WarnContext(ctx, "EXIT executeOne: request failed",
			slog.Int("index", index),
			slog.String("error", err.Error()),
		)
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	slog.DebugContext(ctx, "EXIT executeOne", slog.Int("index", index), slog.Bool("success", resp.Success))
	return resp
}
