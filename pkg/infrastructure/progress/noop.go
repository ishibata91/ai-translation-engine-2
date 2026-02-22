package progress

import "context"

// NoopNotifier は OnProgress を何もせずに無視するデフォルト実装。
// テストや、進捗通知が不要なシナリオで ProgressNotifier の代替として使用する。
type NoopNotifier struct{}

// OnProgress は何もしない。
func (n *NoopNotifier) OnProgress(_ context.Context, _ ProgressEvent) {}

// NewNoopNotifier creates a new NoopNotifier.
func NewNoopNotifier() ProgressNotifier {
	return &NoopNotifier{}
}
