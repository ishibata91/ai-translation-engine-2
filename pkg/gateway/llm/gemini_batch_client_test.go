package llm

import "testing"

func TestParseGeminiBatchStatus_Normalization(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedState BatchState
	}{
		{
			name:          "pending -> queued",
			body:          `{"batch":{"state":"BATCH_STATE_PENDING","batchStats":{"requestCount":"10","successfulRequestCount":"0","failedRequestCount":"0"}}}`,
			expectedState: BatchStateQueued,
		},
		{
			name:          "running -> running",
			body:          `{"batch":{"state":"BATCH_STATE_RUNNING","batchStats":{"requestCount":"10","successfulRequestCount":"3","failedRequestCount":"1"}}}`,
			expectedState: BatchStateRunning,
		},
		{
			name:          "succeeded(all success) -> completed",
			body:          `{"batch":{"state":"BATCH_STATE_SUCCEEDED","batchStats":{"requestCount":"10","successfulRequestCount":"10","failedRequestCount":"0"}}}`,
			expectedState: BatchStateCompleted,
		},
		{
			name:          "succeeded(partial) -> partial_failed",
			body:          `{"batch":{"state":"BATCH_STATE_SUCCEEDED","batchStats":{"requestCount":"10","successfulRequestCount":"8","failedRequestCount":"2"}}}`,
			expectedState: BatchStatePartialFailed,
		},
		{
			name:          "failed -> failed",
			body:          `{"batch":{"state":"BATCH_STATE_FAILED","batchStats":{"requestCount":"10","successfulRequestCount":"0","failedRequestCount":"10"}}}`,
			expectedState: BatchStateFailed,
		},
		{
			name:          "cancelled -> cancelled",
			body:          `{"batch":{"state":"BATCH_STATE_CANCELLED","batchStats":{"requestCount":"10","successfulRequestCount":"2","failedRequestCount":"3"}}}`,
			expectedState: BatchStateCancelled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, err := parseGeminiBatchStatus([]byte(tc.body), "batches/test-1")
			if err != nil {
				t.Fatalf("parseGeminiBatchStatus error: %v", err)
			}
			if status.State != tc.expectedState {
				t.Fatalf("state = %q, want %q", status.State, tc.expectedState)
			}
		})
	}
}

func TestParseGeminiBatchResults_InlinedResponses(t *testing.T) {
	body := `{
		"batch": {
			"output": {
				"inlinedResponses": {
					"inlinedResponses": [
						{
							"metadata": {"req_id": "1"},
							"response": {
								"candidates": [{"content": {"parts": [{"text": "ok"}]}}],
								"usageMetadata": {
									"promptTokenCount": 3,
									"candidatesTokenCount": 4,
									"totalTokenCount": 7
								}
							}
						},
						{
							"metadata": {"req_id": "2"},
							"error": {"message": "bad request"}
						}
					]
				}
			}
		}
	}`

	results, err := parseGeminiBatchResults([]byte(body))
	if err != nil {
		t.Fatalf("parseGeminiBatchResults error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	if !results[0].Success || results[0].Content != "ok" {
		t.Fatalf("first result unexpected: %+v", results[0])
	}
	if results[0].Usage.TotalTokens != 7 {
		t.Fatalf("first result total tokens = %d, want 7", results[0].Usage.TotalTokens)
	}

	if results[1].Success {
		t.Fatalf("second result should fail: %+v", results[1])
	}
	if results[1].Error != "bad request" {
		t.Fatalf("second result error = %q, want %q", results[1].Error, "bad request")
	}
}

func TestExtractGeminiBatchName(t *testing.T) {
	body := []byte(`{"response":{"name":"batches/batch-123"}}`)
	id, err := extractGeminiBatchName(body)
	if err != nil {
		t.Fatalf("extractGeminiBatchName error: %v", err)
	}
	if id != "batches/batch-123" {
		t.Fatalf("id = %q, want %q", id, "batches/batch-123")
	}
}
