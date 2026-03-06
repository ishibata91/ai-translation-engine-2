## Why

現状の `MasterPersona` UI は `pkg` 側に未接続で、開始ボタンを押しても実処理が起動せず何も起こらない。まずはUIと `pkg` を最小構成で接続し、段階的統合の第一段として今回はペルソナリクエスト生成までを成立させる必要がある。

## What Changes

- `MasterPersona` 画面の「開始」操作を `pkg` のペルソナ生成フロー（Phase 1: Propose）へ最小接続し、押下時に処理が起動する状態にする。
- `pkg` 側に、入力データからペルソナ生成リクエストを組み立てて返す境界API（UI呼び出し用）を追加する。
- 既存の log-viewer に、生成結果（リクエスト件数・主要フィールド）を `info` ログとして出力し、UIから `pkg` までの疎通を確認できるようにする。
- 段階的統合方針として、本変更のゴールは「リクエスト生成まで」とし、実際のLLM送信、レスポンス保存、最終確定処理は対象外（No Goals）とする。

## Capabilities

### New Capabilities
- `persona-request-preview`: マスターペルソナ生成開始時に、LLM送信前のリクエスト内容を既存 log-viewer の `info` ログで確認できる機能。

### Modified Capabilities
- `persona`: ペルソナ生成のPhase 1（Propose）を、UIトリガーから実行しログ確認できる統合要件を追加する。

## Impact

- 影響コード: `frontend/src/pages/MasterPersona.tsx`、Wailsバインディング層、`pkg` のペルソナ生成スライス（Propose経路）。
- API/契約: UIから呼び出す「ペルソナリクエスト生成」用の公開メソッドを追加。
- 依存: 既存のWailsイベント/バインディングと `pkg` インターフェースを活用し、新規に非デファクトなライブラリは追加しない。
- DB/外部連携: 本タスクではDBスキーマ変更なし、LLM通信なし。
