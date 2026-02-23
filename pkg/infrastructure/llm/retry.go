package llm

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryableError はリトライ対象のHTTPエラーを表す。
type RetryableError struct {
	StatusCode int
	Message    string
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable HTTP error %d: %s", e.StatusCode, e.Message)
}

// IsRetryableStatusCode は与えられたHTTPステータスコードがリトライ対象かを返す。
// リトライ対象: 429 / 500 / 502 / 503 / 504
// 非リトライ対象: 4xx系（429を除く）
func IsRetryableStatusCode(code int) bool {
	switch code {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// RetryConfig はリトライ動作の設定を保持する。
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// DefaultRetryConfig はデフォルトのリトライ設定を返す。
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}
}

// RetryWithBackoff は fn をリトライポリシーに従って実行する。
// fn が *RetryableError を返した場合のみリトライし、それ以外のエラーは即座に返す。
// ctx がキャンセルされた場合はリトライを中断する。
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	interval := cfg.InitialInterval

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled before attempt %d: %w", attempt+1, err)
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// リトライ対象かチェック
		var retryErr *RetryableError
		if !errors.As(lastErr, &retryErr) {
			// 非リトライエラーは即返す
			return lastErr
		}

		// 最後の試行ではスリープしない
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Exponential Backoff + Jitter
		jitter := time.Duration(rand.Float64() * float64(interval) * 0.2)
		wait := interval + jitter
		if wait > cfg.MaxInterval {
			wait = cfg.MaxInterval
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		case <-time.After(wait):
		}

		interval = time.Duration(float64(interval) * cfg.Multiplier)
		if interval > cfg.MaxInterval {
			interval = cfg.MaxInterval
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", cfg.MaxAttempts, lastErr)
}
