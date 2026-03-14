package llm

import (
	"context"
	"io"
	"log/slog"
	"math"
	"testing"
)

func TestXAIBatchStatusNormalization(t *testing.T) {
	client := &xaiBatchClient{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ctx := context.Background()

	tests := []struct {
		name          string
		body          string
		expectedState BatchState
		expectedProg  float32
	}{
		{
			name:          "num_requests=0 は queued",
			body:          `{"state":{"num_requests":0,"num_pending":0,"num_success":0,"num_error":0,"num_cancelled":0}}`,
			expectedState: "queued",
			expectedProg:  0,
		},
		{
			name:          "num_pending>0 は running",
			body:          `{"state":{"num_requests":10,"num_pending":3,"num_success":5,"num_error":1,"num_cancelled":1}}`,
			expectedState: "running",
			expectedProg:  0.7,
		},
		{
			name:          "all success は completed",
			body:          `{"state":{"num_requests":4,"num_pending":0,"num_success":4,"num_error":0,"num_cancelled":0}}`,
			expectedState: "completed",
			expectedProg:  1,
		},
		{
			name:          "all cancelled は cancelled",
			body:          `{"state":{"num_requests":4,"num_pending":0,"num_success":0,"num_error":0,"num_cancelled":4}}`,
			expectedState: "cancelled",
			expectedProg:  1,
		},
		{
			name:          "all error は failed",
			body:          `{"state":{"num_requests":4,"num_pending":0,"num_success":0,"num_error":4,"num_cancelled":0}}`,
			expectedState: "failed",
			expectedProg:  1,
		},
		{
			name:          "success+error は partial_failed",
			body:          `{"state":{"num_requests":10,"num_pending":0,"num_success":8,"num_error":2,"num_cancelled":0}}`,
			expectedState: "partial_failed",
			expectedProg:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, err := client.parseBatchStatus(ctx, []byte(tc.body), "batch-1")
			if err != nil {
				t.Fatalf("parseBatchStatus error: %v", err)
			}
			if status.State != tc.expectedState {
				t.Fatalf("state = %q, want %q", status.State, tc.expectedState)
			}
			if math.Abs(float64(status.Progress-tc.expectedProg)) > 1e-6 {
				t.Fatalf("progress = %v, want %v", status.Progress, tc.expectedProg)
			}
		})
	}
}

func TestXAIParseResultsPagination(t *testing.T) {
	client := &xaiBatchClient{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	ctx := context.Background()

	body := `{
	  "results": [
	    {
	      "batch_request_id": "req-1",
	      "batch_result": {
	        "response": {
	          "chat_get_completion": {
	            "choices": [{"message": {"content": "ok"}}],
	            "usage": {"prompt_tokens": 1, "completion_tokens": 2, "total_tokens": 3}
	          }
	        }
	      }
	    },
	    {
	      "batch_request_id": "req-2",
	      "error": {"message": "bad request"}
	    }
	  ],
	  "pagination_token": "token-2"
	}`

	results, token, err := client.parseResults(ctx, []byte(body))
	if err != nil {
		t.Fatalf("parseResults error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if !results[0].Success || results[0].Content != "ok" {
		t.Fatalf("first result unexpected: %+v", results[0])
	}
	if results[1].Success || results[1].Error != "bad request" {
		t.Fatalf("second result unexpected: %+v", results[1])
	}
	if token != "token-2" {
		t.Fatalf("token = %q, want %q", token, "token-2")
	}
}
