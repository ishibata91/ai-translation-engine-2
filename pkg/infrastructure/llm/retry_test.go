package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name        string
		statusCodes []int // モックサーバーが順に返すステータスコード
		wantErr     bool
		wantRetries int // 期待する試行回数
	}{
		{
			name:        "正常系: 1回目で成功",
			statusCodes: []int{200},
			wantErr:     false,
			wantRetries: 1,
		},
		{
			name:        "正常系: 429後にリトライして成功",
			statusCodes: []int{429, 429, 200},
			wantErr:     false,
			wantRetries: 3,
		},
		{
			name:        "正常系: 503後にリトライして成功",
			statusCodes: []int{503, 200},
			wantErr:     false,
			wantRetries: 2,
		},
		{
			name:        "正常系: 502後にリトライして成功",
			statusCodes: []int{502, 503, 200},
			wantErr:     false,
			wantRetries: 3,
		},
		{
			name:        "異常系: 429がmaxAttempts回続いてエラー",
			statusCodes: []int{429, 429, 429, 429},
			wantErr:     true,
			wantRetries: 3, // maxAttempts=3 まで
		},
		{
			name:        "異常系: 401は非リトライ対象でエラー",
			statusCodes: []int{401},
			wantErr:     true,
			wantRetries: 1, // リトライしない
		},
		{
			name:        "異常系: 400は非リトライ対象でエラー",
			statusCodes: []int{400},
			wantErr:     true,
			wantRetries: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			idx := 0

			cfg := DefaultRetryConfig()
			cfg.MaxAttempts = 3
			cfg.InitialInterval = 0 // テスト高速化のためインターバルを0に

			err := RetryWithBackoff(context.Background(), cfg, func() error {
				callCount++
				if idx >= len(tc.statusCodes) {
					return nil
				}
				code := tc.statusCodes[idx]
				idx++

				if IsRetryableStatusCode(code) {
					return &RetryableError{StatusCode: code, Message: http.StatusText(code)}
				}
				if code != http.StatusOK {
					// 非リトライエラー（wrap しない）
					return errors.New(http.StatusText(code))
				}
				return nil
			})

			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if callCount != tc.wantRetries {
				t.Errorf("expected %d calls, got %d", tc.wantRetries, callCount)
			}
		})
	}
}

func TestIsRetryableStatusCode(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusTooManyRequests, true},     // 429
		{http.StatusInternalServerError, true}, // 500
		{http.StatusBadGateway, true},          // 502
		{http.StatusServiceUnavailable, true},  // 503
		{http.StatusGatewayTimeout, true},      // 504
		{http.StatusOK, false},                 // 200
		{http.StatusBadRequest, false},         // 400
		{http.StatusUnauthorized, false},       // 401
		{http.StatusForbidden, false},          // 403
		{http.StatusNotFound, false},           // 404
	}

	for _, tc := range tests {
		t.Run(http.StatusText(tc.code), func(t *testing.T) {
			got := IsRetryableStatusCode(tc.code)
			if got != tc.want {
				t.Errorf("IsRetryableStatusCode(%d) = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

// TestRetryContextCancellation は context キャンセル時にリトライが中断されることを確認する。
func TestRetryContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	cfg := DefaultRetryConfig()
	err := RetryWithBackoff(ctx, cfg, func() error {
		return &RetryableError{StatusCode: 429, Message: "rate limit"}
	})

	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

// TestRetryMockHTTPServer は httptest を使ったモックサーバーでのリトライ確認。
func TestRetryMockHTTPServer(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.InitialInterval = 0

	err := RetryWithBackoff(context.Background(), cfg, func() error {
		resp, err := http.Get(server.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if IsRetryableStatusCode(resp.StatusCode) {
			return &RetryableError{StatusCode: resp.StatusCode}
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}
