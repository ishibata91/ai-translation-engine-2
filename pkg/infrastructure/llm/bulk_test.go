package llm

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// --- Mock ---

// mockLLMClient はテスト用のモック LLMClient。
// fn が non-nil の場合、Complete 呼び出しごとに fn を実行する。
// maxParallel は同時並列実行数の最大値を atomic に記録する。
type mockLLMClient struct {
	fn          func(ctx context.Context, index int, req Request) (Response, error)
	callCount   atomic.Int64
	maxParallel atomic.Int64
	current     atomic.Int64
}

func (m *mockLLMClient) Complete(ctx context.Context, req Request) (Response, error) {
	m.callCount.Add(1)
	cur := m.current.Add(1)
	defer m.current.Add(-1)

	for {
		old := m.maxParallel.Load()
		if cur <= old {
			break
		}
		if m.maxParallel.CompareAndSwap(old, cur) {
			break
		}
	}

	if m.fn != nil {
		index := int(m.callCount.Load()) - 1
		return m.fn(ctx, index, req)
	}
	return Response{
		Content: fmt.Sprintf("ok:%s", req.UserPrompt),
		Success: true,
	}, nil
}

func (m *mockLLMClient) StreamComplete(ctx context.Context, req Request) (StreamResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, errors.New("not implemented")
}

func (m *mockLLMClient) HealthCheck(ctx context.Context) error { return nil }

// --- Table-Driven Tests ---

func TestExecuteBulkSync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		numReqs     int
		concurrency int
		// clientFn はreqのUserPromptからインデックスを特定できないため、
		// 呼び出し順のインデックスは使わず req.UserPrompt で判定する。
		clientFn       func(ctx context.Context, index int, req Request) (Response, error)
		ctxTimeout     time.Duration // 0 = no timeout
		wantErrContain string        // 空文字 = error は期待しない
		// wantResults は nil でなければ index ごとに検証する。
		wantResults func(t *testing.T, results []Response)
		// wantMaxParallel > 0 のとき、実測の最大並列数がこの値を超えないことを検証する。
		wantMaxParallel int64
	}{
		{
			name:        "正常: 10件を Concurrency=3 で処理し、入力順序通りに結果が返る",
			numReqs:     10,
			concurrency: 3,
			wantResults: func(t *testing.T, results []Response) {
				for i, res := range results {
					wantContent := fmt.Sprintf("ok:prompt-%d", i)
					if res.Content != wantContent {
						t.Errorf("index %d: got Content=%q, want %q", i, res.Content, wantContent)
					}
					if !res.Success {
						t.Errorf("index %d: expected Success=true", i)
					}
				}
			},
			wantMaxParallel: 3,
		},
		{
			name:        "正常: Concurrency=1 (直列) で処理が完了する",
			numReqs:     5,
			concurrency: 1,
			wantResults: func(t *testing.T, results []Response) {
				if len(results) != 5 {
					t.Fatalf("expected 5 results, got %d", len(results))
				}
				for i, res := range results {
					if !res.Success {
						t.Errorf("index %d: expected Success=true", i)
					}
				}
			},
			wantMaxParallel: 1,
		},
		{
			name:        "正常: Concurrency <= 0 は 1 に補正される",
			numReqs:     3,
			concurrency: 0,
			wantResults: func(t *testing.T, results []Response) {
				if len(results) != 3 {
					t.Fatalf("expected 3 results, got %d", len(results))
				}
			},
			wantMaxParallel: 1,
		},
		{
			name:        "正常: 0件のリクエストは空スライスを返す",
			numReqs:     0,
			concurrency: 5,
			wantResults: func(t *testing.T, results []Response) {
				if len(results) != 0 {
					t.Errorf("expected 0 results, got %d", len(results))
				}
			},
		},
		{
			name:        "PartialFailure: 一部のリクエストがエラーになっても全体は error を返さない",
			numReqs:     5,
			concurrency: 2,
			clientFn: func(ctx context.Context, _ int, req Request) (Response, error) {
				// UserPrompt が "prompt-2" の場合のみエラーにする
				if req.UserPrompt == "prompt-2" {
					return Response{}, errors.New("rate limit exceeded")
				}
				return Response{Content: "ok:" + req.UserPrompt, Success: true}, nil
			},
			wantResults: func(t *testing.T, results []Response) {
				for i, res := range results {
					if results[i].Success && results[i].Content != fmt.Sprintf("ok:prompt-%d", i) {
						t.Errorf("index %d: unexpected content %q", i, res.Content)
					}
				}
				// index 2 は失敗のはずだが、全体エラーではなく Response.Success=false
				failCount := 0
				for _, res := range results {
					if !res.Success {
						failCount++
						if res.Error == "" {
							t.Error("failed Response should have non-empty Error field")
						}
					}
				}
				if failCount != 1 {
					t.Errorf("expected exactly 1 failed result, got %d", failCount)
				}
			},
		},
		{
			name:        "PartialFailure: 全リクエストがエラーでも関数は error を返さない",
			numReqs:     4,
			concurrency: 2,
			clientFn: func(ctx context.Context, _ int, req Request) (Response, error) {
				return Response{}, errors.New("always fail")
			},
			wantResults: func(t *testing.T, results []Response) {
				for i, res := range results {
					if res.Success {
						t.Errorf("index %d: expected Success=false", i)
					}
					if res.Error == "" {
						t.Errorf("index %d: expected non-empty Error", i)
					}
				}
			},
		},
		{
			name:           "ContextCancellation: タイムアウトしたコンテキストは ctx.Err() を返す",
			numReqs:        10,
			concurrency:    2,
			ctxTimeout:     30 * time.Millisecond,
			wantErrContain: "context",
			clientFn: func(ctx context.Context, _ int, req Request) (Response, error) {
				select {
				case <-ctx.Done():
					return Response{}, ctx.Err()
				case <-time.After(500 * time.Millisecond):
					return Response{Content: "ok", Success: true}, nil
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tc.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
				defer cancel()
			}

			client := &mockLLMClient{fn: tc.clientFn}

			reqs := make([]Request, tc.numReqs)
			for i := range reqs {
				reqs[i] = Request{UserPrompt: fmt.Sprintf("prompt-%d", i)}
			}

			results, err := ExecuteBulkSync(ctx, client, reqs, tc.concurrency)

			// エラー検証
			if tc.wantErrContain != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrContain)
				}
				// エラーが context 系であることを確認
				if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
					t.Errorf("expected context error, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 件数検証
			if len(results) != tc.numReqs {
				t.Fatalf("expected %d results, got %d", tc.numReqs, len(results))
			}

			// 結果内容の検証
			if tc.wantResults != nil {
				tc.wantResults(t, results)
			}

			// 並列数制限の検証
			if tc.wantMaxParallel > 0 {
				if got := client.maxParallel.Load(); got > tc.wantMaxParallel {
					t.Errorf("max parallel exceeded concurrency: got %d, want <= %d", got, tc.wantMaxParallel)
				}
			}
		})
	}
}
