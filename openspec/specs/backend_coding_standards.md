# バックエンド標準コーディング規約

本規約は `openspec/specs/architecture.md` と `openspec/specs/standard_test_spec.md` に準拠し、`pkg/` を中心とするバックエンド実装とそのレビュー基準を固定する。

## 1. 適用範囲

- 対象は `pkg/**` と `cmd/**` を中心とする。
- `app.go` / `main.go` など Wails エントリポイントは参考準拠とし、標準品質ゲートの必須判定対象には含めない。
- フロントエンド固有の実装規約は `openspec/specs/frontend_architecture.md` を参照する。
- 本規約における `MUST` は必須、`SHOULD` は正当な理由がある場合のみ例外を認める推奨事項とする。

## 2. 共通規約

### 2.1 MUST

- 命名は Go 標準の慣例に従い、略語は `ID`, `URL`, `HTTP` のように一貫した大文字規則で扱うこと。
- 返却する `error` は呼び出し元で原因を追跡できるように `fmt.Errorf("...: %w", err)` などで文脈を付与すること。
- `context.Context` は公開 Contract メソッドの第一引数で受け取り、内部処理へ必ず伝播すること。
- ログ出力は `slog.*Context` を使用し、構造化フィールドで slice 名・入力件数・識別子などの分析可能な情報を記録すること。
- 公開メソッドは 1 つの責務に保ち、複雑な分岐や処理列は同一ファイル内のプライベートメソッドへ分割すること。
- 公開型、公開関数、公開メソッドには doc コメントを付与すること。
- テストは `openspec/specs/standard_test_spec.md` に整合し、Table-Driven Test を主軸にすること。
- テストおよび本番コードのデバッグは構造化ログ前提で行い、問題解析のために `context.Context` を途切れさせないこと。

### 2.2 SHOULD

- 依存注入はコンストラクタで行い、具象実装への直接依存を避ける。
- `context.Context` を受け取らない内部ヘルパーは、副作用がない純粋関数に限定する。
- ログメッセージは機械可読な識別子を優先し、日本語説明文は必要最小限に留める。
- テストは in-memory SQLite または局所モックを優先し、外部環境依存を持ち込まない。

## 3. リポジトリ固有規約

### 3.1 MUST

- Interface-First AIDD に従い、変更の起点は Contract・Spec・Plan とし、実装詳細への直接依存を増やさないこと。
- Vertical Slice の自律性を優先し、業務ロジックや DTO を安易に共通化しないこと。
- 共有 DTO を `pkg/domain` のような横断モデルとして導入せず、各 slice が自身の入力 DTO / 出力 DTO を持つこと。
- Slice 間のデータ変換は slice 内ではなく、Pipeline / Orchestrator / Mapper 層で担うこと。
- Slice は他 slice の具象実装を参照せず、必要な連携は Contract 越しに行うこと。

### 3.2 SHOULD

- 共通化が必要な場合は、それが技術的関心事か不変の真理かを明示してから導入する。
- 1 ファイル内で処理フローが追えるように、外部ファイル分割よりもプライベートメソッド分割を優先する。

## 4. 品質ゲート

- 標準の必須ツールは `golangci-lint`、`goimports`、`goleak`、`go test ./pkg/...` とする。
- `golangci-lint` では少なくとも `staticcheck`、`govet`、`errcheck`、`revive`、`gosec` を有効化する。
- `govulncheck` は依存更新時またはリリース前の任意実行とし、日常開発の必須ブロック条件には含めない。
- 個人開発向けの標準運用として、AI 開発中はローカルで `backend:fmt` / `backend:lint` / `backend:test` / `backend:check` を高頻度に回す。
- section 4 完了前の暫定運用として、自動ブロック条件はローカルの整形差分なしと `go test ./pkg/...` 成功を基準とし、`backend:lint` は毎回実行して結果を可視化する。

### 4.1 標準コマンド

- `npm run backend:fmt`
- `npm run backend:lint`
- `npm run backend:test`
- `npm run backend:check`
- `npm run backend:watch`
- `npm run backend:watch:lint`
- `npm run backend:vuln`

上記コマンドはすべて `go run ./tools/backendquality ...` を通して同一設定で実行する。
`backend:check` はローカルの必須確認用であり、整形差分検出と `go test ./pkg/...` を実行する。
`backend:watch` は `pkg/**`、`cmd/**`、品質設定ファイルの変更をポーリング監視し、変更検知のたびに `backend:check` を再実行する。
`backend:watch:lint` は上記に加えて `backend:lint` も再実行する。

## 5. レビューチェックスタイル

### 5.1 MUST 違反

- `context.Context` 未伝播、error wrap 欠落、構造化ログ逸脱、Contract 境界破壊、`backend:check` の失敗は差し戻し対象とする。
- `backend:lint` の新規違反を持ち込む変更は差し戻し対象とする。
- MUST 違反を許容する場合は、PR 本文に理由・影響範囲・解消予定を明記すること。

### 5.2 SHOULD 違反

- SHOULD 違反は原則修正対象だが、スコープ超過や既存コードとの整合で見送る場合はレビューコメントに判断理由を残す。
- SHOULD を継続的に見送る場合は、別タスク化または OpenSpec 変更へ積み残しを明記する。

### 5.3 例外運用

- 例外は一時的であることを前提とし、恒久化しない。
- 例外を導入する場合は「なぜ必要か」「いつ解消するか」「何で検知するか」を PR または関連タスクへ残す。

## 6. レビュー観点

- Contract / DTO / Mapper の責務境界が `architecture.md` に一致しているか。
- 公開 API に doc コメントがあり、責務が 1 つに保たれているか。
- `error` に十分な文脈が付与され、`context.Context` が公開入口から末端まで流れているか。
- ログが `slog.*Context` で出力され、解析に必要なキーを持っているか。
- テストが Table-Driven を基調とし、必要な並行処理テストには `goleak` が適用されているか。
- ローカルで `backend:check` を満たし、`backend:lint` の結果と必要に応じた `backend:vuln` 実行有無が説明されているか。
