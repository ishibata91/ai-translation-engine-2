package llm

import (
	"math"
	"testing"
)

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
							"metadata": {"queue_job_id": "job-1"},
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
							"metadata": {"queue_job_id": "job-2"},
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
	if got := results[0].Metadata[BatchMetadataQueueJobIDKey]; got != "job-1" {
		t.Fatalf("first result queue_job_id = %v, want job-1", got)
	}
	if got := results[1].Metadata[BatchMetadataQueueJobIDKey]; got != "job-2" {
		t.Fatalf("second result queue_job_id = %v, want job-2", got)
	}
}

func TestGeminiBatchRequestMetadata_RequiresQueueJobID(t *testing.T) {
	_, err := buildGeminiBatchRequestMetadata(map[string]interface{}{"foo": "bar"}, 0)
	if err == nil {
		t.Fatalf("expected queue_job_id validation error")
	}

	metadata, err := buildGeminiBatchRequestMetadata(map[string]interface{}{
		BatchMetadataQueueJobIDKey:      "job-10",
		BatchMetadataQueueRequestSeqKey: 7,
	}, 1)
	if err != nil {
		t.Fatalf("buildGeminiBatchRequestMetadata failed: %v", err)
	}
	if metadata[BatchMetadataQueueJobIDKey] != "job-10" {
		t.Fatalf("queue_job_id = %v, want job-10", metadata[BatchMetadataQueueJobIDKey])
	}
}

func TestParseGeminiBatchStatus_MetadataPayload(t *testing.T) {
	body := `{
		"name": "batches/zyy9gq671p0n1rme102glbhi6b2jgey1lm2u",
		"metadata": {
			"@type": "type.googleapis.com/google.ai.generativelanguage.v1main.GenerateContentBatch",
			"model": "models/gemini-3.1-flash-lite-preview",
			"displayName": "batch-1773512289",
			"createTime": "2026-03-14T18:18:10.082996104Z",
			"updateTime": "2026-03-14T18:18:10.082996104Z",
			"batchStats": {
				"requestCount": "31",
				"pendingRequestCount": "31"
			},
			"state": "BATCH_STATE_PENDING",
			"name": "batches/zyy9gq671p0n1rme102glbhi6b2jgey1lm2u"
		}
	}`

	status, err := parseGeminiBatchStatus([]byte(body), "batches/zyy9gq671p0n1rme102glbhi6b2jgey1lm2u")
	if err != nil {
		t.Fatalf("parseGeminiBatchStatus error: %v", err)
	}
	if status.State != BatchStateQueued {
		t.Fatalf("state = %q, want %q", status.State, BatchStateQueued)
	}
	if !almostEqualFloat32(status.Progress, 0) {
		t.Fatalf("progress = %v, want 0", status.Progress)
	}
}

func TestParseGeminiBatchStatus_ProgressFromPendingCount(t *testing.T) {
	body := `{"metadata":{"state":"BATCH_STATE_RUNNING","batchStats":{"requestCount":"10","pendingRequestCount":"4"}}}`

	status, err := parseGeminiBatchStatus([]byte(body), "batches/test-2")
	if err != nil {
		t.Fatalf("parseGeminiBatchStatus error: %v", err)
	}
	if status.State != BatchStateRunning {
		t.Fatalf("state = %q, want %q", status.State, BatchStateRunning)
	}
	if !almostEqualFloat32(status.Progress, 0.6) {
		t.Fatalf("progress = %v, want 0.6", status.Progress)
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

func almostEqualFloat32(a, b float32) bool {
	const epsilon = 0.0001
	return math.Abs(float64(a-b)) < epsilon
}
