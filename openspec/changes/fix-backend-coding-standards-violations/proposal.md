## Why

`pkg/**` を走査したところ、`backend_coding_standards.md` の MUST / SHOULD に対して、lint では検知されない `context.Context` 未伝播、error wrap 欠落、構造化ログの不統一、失敗の握りつぶしが残っていた。現状のままでは障害解析と再開系挙動の追跡が難しく、今後のバックエンド修正でも同種の逸脱を再生産しやすいため、この段階で修正対象を OpenSpec change として明文化する必要がある。

## What Changes

- `pkg/task`, `pkg/workflow`, `pkg/config`, `pkg/pipeline` を中心に、`context.Context` を公開入口から内部処理まで途切れさせない方針へ修正する。
- package 境界をまたぐ `error` 返却に文脈付き wrap を追加し、`task_id`, `process_id`, `namespace` などの識別子を追えるようにする。
- `slog.*Context` を基準にログ出力をそろえ、機械可読なメッセージ識別子と固定キーへ寄せる。
- `_ = ...` や `return nil` による失敗の握りつぶしを見直し、許容する箇所と許容しない箇所を整理する。
- 責務過多な公開メソッドは、同一ファイル内 private method 抽出で分割し、レビューしやすい形に整える。
- change 配下の `violation-targets.md` を調査インベントリとして維持し、実装タスクの入力に使う。

## Capabilities

### New Capabilities

- なし: 既存のバックエンド仕様に新しい capability は追加せず、実装を既存規約へ適合させる change とする

### Modified Capabilities

- なし: 既存 spec の REQUIREMENTS 自体は変更せず、`openspec/specs/backend_coding_standards.md` に対する実装準拠を進める

## Impact

- 影響コード: `pkg/task`, `pkg/workflow`, `pkg/config`, `pkg/pipeline` 配下の公開メソッド、store 呼び出し、ログ出力、resume/cancel 周辺処理
- API / UI 影響: Wails バインディングの表向き API 形状は維持する想定だが、内部の `context.Context` 伝播とエラー内容は変わる可能性がある
- 依存関係: 新規ライブラリ追加は想定しない。既存の `context`, `fmt`, `log/slog`, `telemetry` 利用の整理で対応する
- 品質確認: `backend:lint:file -> 修正 -> 再実行 -> lint:backend` と関連 Go テストの再確認が必要になる
