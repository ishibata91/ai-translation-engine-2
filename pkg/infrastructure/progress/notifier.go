package progress

import "context"

const (
	StatusInProgress = "IN_PROGRESS"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

// ProgressEvent はドメイン・インフラに依存しない汎用進捗イベント型。
type ProgressEvent struct {
	CorrelationID string // 処理のまとまりを識別するID (UUID等)。UI側での表示単位となる
	Total         int    // 総件数（不明な場合は 0）
	Completed     int    // 完了件数
	Failed        int    // 失敗件数
	Status        string // "IN_PROGRESS" / "COMPLETED" / "FAILED"
	Message       string // ユーザーに表示する進捗メッセージ
}

// ProgressNotifier は進捗通知の送信先を抽象化するインターフェース。
type ProgressNotifier interface {
	OnProgress(ctx context.Context, event ProgressEvent)
}
