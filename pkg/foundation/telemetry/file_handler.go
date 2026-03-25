package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogRetentionDays はログファイルの保持日数。この日数を超えたファイルは削除される。
const LogRetentionDays = 7

// logDir はリポジトリルート相対のログディレクトリ名。実行バイナリの位置に依存しない設計にするため
// ProvideFileHandler 側で実行ファイルパスから解決する。
const logDir = "logs"

// fileHandler は slog.Handler を実装し、日付ごとの JSON Lines ファイルへ書き込む。
// ファイルは logs/YYYY-MM-DD.jsonl として保存される。
type fileHandler struct {
	mu       sync.Mutex
	baseDir  string       // ログディレクトリの絶対パス
	next     slog.Handler // フォールバック先（コンソール等）
	preAttrs []slog.Attr

	// 現在開いているファイルの日付とファイルポインタ
	currentDate string
	currentFile *os.File
	currentJSON *slog.JSONHandler
}

// newFileHandler は baseDir にログを書くファイルハンドラーを生成する。
// next ハンドラーにもログを流す（マルチキャスト用）。
func newFileHandler(baseDir string, next slog.Handler) (*fileHandler, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("telemetry: failed to create log dir %q: %w", baseDir, err)
	}
	h := &fileHandler{
		baseDir: baseDir,
		next:    next,
	}
	return h, nil
}

// openForToday はログファイルを今日の日付で開く。
// すでに開いている場合は既存ファイルを再利用する。mu がロック済みの状態で呼ぶこと。
func (h *fileHandler) openForToday() error {
	today := time.Now().Format("2006-01-02")
	if today == h.currentDate && h.currentFile != nil {
		return nil
	}

	// 日が変わった or 初回オープン
	if h.currentFile != nil {
		_ = h.currentFile.Close()
		h.currentFile = nil
		h.currentJSON = nil
	}

	path := filepath.Join(h.baseDir, today+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("telemetry: failed to open log file %q: %w", path, err)
	}

	h.currentDate = today
	h.currentFile = f
	h.currentJSON = slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	// 事前属性を再付与
	if len(h.preAttrs) > 0 {
		h.currentJSON = h.currentJSON.WithAttrs(h.preAttrs).(*slog.JSONHandler)
	}
	return nil
}

// Close は現在開いているログファイルを閉じる。アプリ終了時に呼ぶこと。
func (h *fileHandler) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.currentFile != nil {
		_ = h.currentFile.Close()
		h.currentFile = nil
		h.currentJSON = nil
	}
}

func (h *fileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *fileHandler) Handle(ctx context.Context, r slog.Record) error {
	// コンソール等の下流ハンドラーに渡す
	if err := h.next.Handle(ctx, r); err != nil {
		return fmt.Errorf("fileHandler delegate failed: %w", err)
	}

	// ファイルに書き込む
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.openForToday(); err != nil {
		// ファイルオープン失敗はログ欠落として握りつぶす（無限ループ防止）
		return nil
	}

	return h.currentJSON.Handle(ctx, r)
}

func (h *fileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	newPreAttrs := append(append([]slog.Attr{}, h.preAttrs...), attrs...)
	h.mu.Unlock()

	return &fileHandler{
		baseDir:  h.baseDir,
		next:     h.next.WithAttrs(attrs),
		preAttrs: newPreAttrs,
	}
}

func (h *fileHandler) WithGroup(name string) slog.Handler {
	return &fileHandler{
		baseDir:  h.baseDir,
		next:     h.next.WithGroup(name),
		preAttrs: h.preAttrs,
	}
}

// ---- ログ掃除 ---------------------------------------------------------------

// StartLogCleaner はバックグラウンドで定期的に古いログを削除するゴルーチンを起動する。
// ctx がキャンセルされると停止する。
// retentionDays 日より古い .jsonl ファイルを削除する。
func StartLogCleaner(ctx context.Context, baseDir string, retentionDays int) {
	go runLogCleaner(ctx, baseDir, retentionDays)
}

// runLogCleaner は起動時に即時掃除を実施し、その後は 24 時間ごとに掃除する。
func runLogCleaner(ctx context.Context, baseDir string, retentionDays int) {
	// 起動時に即実行
	cleanOldLogs(baseDir, retentionDays)

	// 次の 0 時まで待機してから、24 時間ごとに繰り返す
	ticker := time.NewTicker(nextMidnightDuration())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanOldLogs(baseDir, retentionDays)
			// 次の 24 時間タイマーに更新
			ticker.Reset(24 * time.Hour)
		}
	}
}

// nextMidnightDuration は現在時刻から次の 0 時 0 分 0 秒までの duration を返す。
func nextMidnightDuration() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	d := time.Until(next)
	if d <= 0 {
		d = 24 * time.Hour
	}
	return d
}

// cleanOldLogs は baseDir 内にある retentionDays 日より古い *.jsonl ファイルを削除する。
func cleanOldLogs(baseDir string, retentionDays int) {
	threshold := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		// ディレクトリが存在しない場合等は無視
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".jsonl" {
			continue
		}
		// ファイル名から日付をパース（YYYY-MM-DD.jsonl）
		dateStr := name[:len(name)-len(".jsonl")]
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// パースできない名前のファイルは触らない
			continue
		}
		if t.Before(threshold) {
			_ = os.Remove(filepath.Join(baseDir, name))
		}
	}
}

// resolveLogDir は実行ファイルの位置からリポジトリルートの logs ディレクトリパスを解決する。
// 開発時（wails dev）は実行カレントディレクトリ = リポジトリルートであることが多い。
// 本番ビルドでも同様にバイナリ隣の logs/ を使う。
func resolveLogDir() string {
	// カレントディレクトリ = リポジトリルートを優先
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, logDir)
	}
	// fallback: 実行バイナリ隣
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), logDir)
	}
	return logDir
}
