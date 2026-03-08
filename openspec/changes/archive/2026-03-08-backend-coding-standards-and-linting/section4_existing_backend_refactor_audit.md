# セクション4 既存バックエンド棚卸し

## 目的

`backend-coding-standards-and-linting` の 4.1-4.3 向けに、既存バックエンドの主要違反を優先度順で整理し、即時修正対象と段階導入対象を分離する。

## 棚卸し結果

### P0: 即時修正済み

- `errcheck` の未処理エラー
  - テストコードの `WriteString` / `Set` / `SubmitJobs` / `SaveState`
  - 実装コードの `Discard` / `Rollback` / `UpdateJob` / JSON encode
- `staticcheck` の `nil Context` 渡し
  - `pkg/task/bridge_integration_test.go`
- `revive` の命名違反
  - `pkg/task/manager.go` の `taskId` を `taskID` へ修正
- `unused` / `gosimple`
  - `pkg/infrastructure/telemetry/span.go`
  - `pkg/infrastructure/telemetry/telemetry_integration_test.go`
- `gosec` のファイル権限
  - `pkg/export/exporter.go`
  - `pkg/parser/test/loader_test.go`
- 非推奨 API
  - `pkg/export/exporter_test.go` の `ioutil` を `os` へ置換

### P1: 段階導入対象として一時除外

- `pkg/dictionary/store.go`
  - `gosec G202`
  - 動的検索条件の組み立てが原因。現状はプレースホルダ引数で値は束縛しているため、直近のブロック条件からは除外する。
- `pkg/infrastructure/llm/retry.go`
  - `gosec G404`
  - ジッタ用途の擬似乱数であり、暗号用途ではないため必須ゲートからは一時除外する。
- `pkg/infrastructure/telemetry/provider.go`
  - `revive unexported-return`
  - Wails ハンドラの初期化都合で unexported concrete type を返している。DI 境界再設計時に再対応する。
- `pkg/translator/provider.go`
  - `revive unexported-return`
  - Wire provider と既存 concrete 実装の組み合わせによるもの。 translator slice の DI 整理時に対応する。
- `pkg/translator/persistence.go`
  - `revive unexported-return`
  - 上記と同根。

## 是正計画

1. P0 は本セッションで修正し、`backend:lint` のブロッカーから除去する。
2. P1 は `.golangci.yml` の `exclude-rules` に明示し、例外を局所化する。
3. 次フェーズで DI 境界と SQL 組み立てを再設計し、除外ルールを段階的に削除する。

## 4.3 運用方針

- 現時点の必須 fail 条件は `errcheck` / `staticcheck` / `govet` / `revive` / `gosec` を維持する。
- ただし既知の P1 例外はファイル単位で限定除外し、lint 全体は安定実行可能な状態にする。
- 新規コードで同種の例外を追加する場合は、OpenSpec または PR で理由を明記する。
