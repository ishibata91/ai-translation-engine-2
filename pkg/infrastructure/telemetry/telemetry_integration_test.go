package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
)

// ---- テスト用ヘルパー --------------------------------------------------------

// newBufLogger は出力先を buf に向けた otelHandler ラップのテスト用 slog.Logger を返す。
// グローバル属性を付与し、otelHandler 経由でコンテキスト属性もマージされる。
func newBufLogger(buf *bytes.Buffer) *slog.Logger {
	baseHandler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{
		slog.String("env", "test"),
		slog.String("app_version", "v0.0.1"),
		slog.String("service_name", "test-service"),
	})
	return slog.New(&otelHandler{next: baseHandler})
}

// parseLastLog は buf の全行を読み、最後の JSON オブジェクトを返す。
func parseLastLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	var result map[string]any
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("JSON parse error: %v", err)
		}
		result = m
	}
	return result
}

// ---- Scenario 1: Global Context (執行環境キーの付与) -------------------------

func TestGlobalContextKeys(t *testing.T) {
	var buf bytes.Buffer
	logger := newBufLogger(&buf)

	logger.InfoContext(context.Background(), "test message")

	log := parseLastLog(t, &buf)
	for _, key := range []string{"env", "app_version", "service_name"} {
		if _, ok := log[key]; !ok {
			t.Errorf("expected key %q in log output. log: %v", key, log)
		}
	}
}

// ---- Scenario 2: Semantic Action (セマンティクスキーの付与) ------------------

func TestWithAction(t *testing.T) {
	var buf bytes.Buffer
	logger := newBufLogger(&buf)

	ctx := context.Background()
	ctx = WithAction(ctx, ActionImport, ResourceDictionary, "dict-001")

	logger.InfoContext(ctx, "action test")

	log := parseLastLog(t, &buf)
	checks := map[string]string{
		"action":        "import",
		"resource_type": "dictionary",
		"resource_id":   "dict-001",
	}
	for k, v := range checks {
		got, ok := log[k]
		if !ok {
			t.Errorf("key %q not found in log. log: %v", k, log)
			continue
		}
		if got != v {
			t.Errorf("key %q: expected %q, got %v", k, v, got)
		}
	}
}

// TestWithAttrs_Isolation は WithAttrs が既存コンテキストの属性スライスを
// 変更しないことを確認する（副作用なし）。
func TestWithAttrs_Isolation(t *testing.T) {
	ctx := context.Background()
	ctx1 := WithAction(ctx, ActionImport, ResourceDictionary, "d-1")
	// ctx2 はさらに属性を追加
	ctx2 := WithAction(ctx1, ActionTranslate, ResourceEntry, "e-2")

	attrs1 := attrsFromContext(ctx1)
	attrs2 := attrsFromContext(ctx2)

	if len(attrs1) != 3 {
		t.Errorf("ctx1 should have 3 attrs (action, resource_type, resource_id), got %d: %v", len(attrs1), attrs1)
	}
	if len(attrs2) != 6 {
		t.Errorf("ctx2 should have 6 attrs (3+3), got %d: %v", len(attrs2), attrs2)
	}
}

// ---- Scenario 3: Performance Tracking (パフォーマンス計測) -------------------

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	// defer パターンで panic や error が起きないことを確認
	end := StartSpan(ctx, ActionImport)
	end()
}

// TestStartSpan_LogsOutput は StartSpan が span.start と span.end の
// ログを出力することを確認する。
func TestStartSpan_LogsOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := newBufLogger(&buf)

	// StartSpan の内部は slog.Default() を使うため、テスト用ロガーをデフォルトに設定
	orig := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(orig)

	ctx := context.Background()
	end := StartSpan(ctx, ActionImport)
	end()

	output := buf.String()
	if output == "" {
		t.Error("expected log output from StartSpan, got empty")
	}
	// span.start と span.end の 2 行が出力されることを確認
	dec := json.NewDecoder(&buf)
	var count int
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			break
		}
		count++
	}
	if count < 2 {
		t.Errorf("expected at least 2 log lines from StartSpan, got %d. output: %s", count, output)
	}
}

// ---- Scenario 4: Error Context (エラー解決キーへの対応) ----------------------

func TestErrorAttrs(t *testing.T) {
	err := errors.New("something went wrong")
	attrs := ErrorAttrs(err)

	keys := map[string]bool{}
	for _, a := range attrs {
		keys[a.Key] = true
	}

	for _, expected := range []string{"error_code", "exception_class", "stack_trace", "error_message"} {
		if !keys[expected] {
			t.Errorf("expected key %q in ErrorAttrs output. attrs: %v", expected, attrs)
		}
	}
}

func TestErrorAttrs_Nil(t *testing.T) {
	attrs := ErrorAttrs(nil)
	if len(attrs) != 0 {
		t.Errorf("expected empty attrs for nil error, got: %v", attrs)
	}
}

// TestErrorAttrs_InLog は ErrorAttrs がログ出力に組み込めることを確認する。
func TestErrorAttrs_InLog(t *testing.T) {
	var buf bytes.Buffer
	logger := newBufLogger(&buf)

	ctx := context.Background()
	err := errors.New("test error")
	errAttrs := ErrorAttrs(err)

	// slog.Attr スライスを ...any に変換してログ出力
	args := make([]any, len(errAttrs))
	for i, a := range errAttrs {
		args[i] = a
	}
	logger.ErrorContext(ctx, "operation failed", args...)

	log := parseLastLog(t, &buf)
	for _, key := range []string{"error_code", "exception_class"} {
		if _, ok := log[key]; !ok {
			t.Errorf("expected key %q in log output. log: %v", key, log)
		}
	}
}
