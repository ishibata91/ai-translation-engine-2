package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// ---- StartSpan: パフォーマンス計測 -----------------------------------------

// Span は実行中のスパン情報を保持する。
type Span struct {
	ctx       context.Context
	action    ActionType
	startedAt time.Time
	logger    *slog.Logger
}

// StartSpan はスパンを開始し、defer で呼び出す終了関数を返す。
// 終了関数が呼ばれると duration_ms 付きの Exit ログが出力される。
//
// 推奨利用方法:
//
//	defer telemetry.StartSpan(ctx, telemetry.ActionImport)()
//
// または span.End() を明示的に呼ぶ場合:
//
//	span := telemetry.StartSpan(ctx, telemetry.ActionImport)
//	defer span()
func StartSpan(ctx context.Context, action ActionType) func() {
	logger := slog.Default()
	start := time.Now()

	// 開始ログ
	logger.InfoContext(ctx, "span.start",
		slog.String("action", string(action)),
	)

	return func() {
		elapsed := time.Since(start)
		ms := float64(elapsed.Microseconds()) / 1000.0

		logger.InfoContext(ctx, "span.end",
			slog.String("action", string(action)),
			slog.Float64("duration_ms", ms),
		)
	}
}

// ---- ErrorAttrs: エラー属性生成 --------------------------------------------

// ErrorAttrs は error オブジェクトから構造化ログ用の slog.Attr スライスを生成する。
// 以下のキーを標準的なフォーマットで出力する:
//   - error_code    : HTTP ステータスコードや内部エラーコード（interface から取得、なければ "UNKNOWN"）
//   - exception_class: エラーの型名
//   - stack_trace   : スタックトレース文字列
//
// 利用例:
//
//	logger.ErrorContext(ctx, "operation failed", telemetry.ErrorAttrs(err)...)
func ErrorAttrs(err error) []slog.Attr {
	if err == nil {
		return nil
	}

	attrs := []slog.Attr{
		slog.String("error_code", extractErrorCode(err)),
		slog.String("exception_class", extractExceptionClass(err)),
		slog.String("stack_trace", captureStackTrace()),
		slog.String("error_message", err.Error()),
	}
	return attrs
}

// ---- 内部ヘルパー ------------------------------------------------------------

// errCoder はエラーコードを持つ error インターフェース。
// カスタムエラー型がこのインターフェースを実装していれば error_code を取得できる。
type errCoder interface {
	ErrorCode() string
}

// extractErrorCode はエラーからエラーコードを取得する。
// errCoder インターフェースを実装していない場合は "UNKNOWN" を返す。
func extractErrorCode(err error) string {
	var coder errCoder
	if errors.As(err, &coder) {
		return coder.ErrorCode()
	}
	return "UNKNOWN"
}

// extractExceptionClass はエラーの型名を返す。
// エラー型を unwrap しながら最初に見つかった具体型の名前を返す。
func extractExceptionClass(err error) string {
	if err == nil {
		return ""
	}
	// fmt.Sprintf の %T は型名を "pkg.TypeName" の形式で返す
	typeName := fmt.Sprintf("%T", err)
	// パッケージパスを除いて型名のみ返す
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		return typeName[idx+1:]
	}
	return typeName
}

// captureStackTrace は現在のゴルーチンのスタックトレース文字列を返す。
// telemetry パッケージ自身のフレームは除外し、呼び出し元から始まるトレースを返す。
func captureStackTrace() string {
	const maxFrames = 32
	pcs := make([]uintptr, maxFrames)
	// skip: runtime.Callers(1) + captureStackTrace + ErrorAttrs の3フレームをスキップ
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder
	for {
		frame, more := frames.Next()
		// telemetry パッケージ内部のフレームはスキップ
		if strings.Contains(frame.Function, "telemetry.") {
			if !more {
				break
			}
			continue
		}
		fmt.Fprintf(&sb, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return sb.String()
}
