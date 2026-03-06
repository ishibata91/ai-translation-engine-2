## 1. タスク経路の統合（task.Bridge）

- [ ] 1.1 `task.Bridge` にマスターペルソナ生成タスク起動API（Phase 1専用）を追加する
- [ ] 1.2 `task.Manager` 側でタスクID発行と状態遷移（`Pending` -> `Running` -> `REQUEST_GENERATED` / `Failed`）を実装する
- [ ] 1.3 `pkg/persona` の `PreparePrompts` 呼び出しをタスク実行関数に接続し、同一タスクIDで追跡可能にする

## 2. UI接続（MasterPersona）

- [ ] 2.1 `frontend/src/pages/MasterPersona.tsx` の開始ボタンを新しいタスク起動APIに接続する
- [ ] 2.2 タスク更新イベントを購読し、生成中表示の開始/解除と完了ステータス表示を実装する
- [ ] 2.3 成功時に件数・サンプルの最小サマリのみUI状態として保持し、全量リクエストは保持しない

## 3. ログ連携（既存log-viewer利用）

- [ ] 3.1 生成成功時に `persona.requests.generated` の `info` ログを出力する（`request_count`、`npc_count`、`task_id` を含む）
- [ ] 3.2 生成失敗時に `persona.requests.failed` の `error` ログを出力する（`task_id`、失敗理由を含む）
- [ ] 3.3 `wails dev` で log-viewer のレベルを `INFO` 以上に設定した際にログが確認できることを手順化する

## 4. スコアリング変更（probe廃止）

- [ ] 4.1 `pkg/persona` のスコア計算から `probe` 依存処理を削除する
- [ ] 4.2 英語ダイアログで大文字フレーズ出現率を算出するロジックを追加する
- [ ] 4.3 日本語ダイアログ判定時は大文字フレーズ率スコアをスキップし、他特徴量のみで順位付けする

## 5. テストと受け入れ確認

- [ ] 5.1 `pkg/persona` のTable-Driven Testを更新し、英語/日本語のスコアリング分岐を検証する
- [ ] 5.2 タスク状態遷移とログ属性（`task_id` 等）を確認する統合テストまたは検証ケースを追加する
- [ ] 5.3 受け入れ確認として「開始押下 -> REQUEST_GENERATED -> log-viewerでinfo確認」を実施し、結果を記録する
