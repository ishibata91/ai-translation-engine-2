# Change Review Checklist Template

このファイルは、`openspec/changes/<change>/review.md` を書くときのテンプレートである。
ここには change 固有の観点だけを書く。共通観点は `openspec/review_standard.md` を前提とする。

## 1. Scope

- この change で主に確認したい領域: package 境界と外部 I/O 境界の error wrap 不足検出
- 影響を強く受ける package / feature: `tools/backendquality`, `pkg/**`, lint 実行導線
- 特に見落としやすい境界: `return err`, `%w` 欠落, cleanup と本流 error path の区別

## 2. Backend-Specific Checks

必要な場合のみ記入する。

- [ ] Contract / DTO / Mapper / Orchestrator の責務境界が change の意図どおりか
- [ ] `context.Context` の伝播が新規コードでも維持されているか
- [ ] `pkg/infrastructure/telemetry` を使うべき処理で独自 logging を増やしていないか
- [ ] 外部 I/O、重要分岐、状態変化、異常系に AI 解析可能な構造化ログがあるか
- [ ] 追加した error path に文脈付き error wrap があるか
- [ ] 主要シナリオと失敗系をカバーする Go テストがあるか

## 3. Frontend-Specific Checks

必要な場合のみ記入する。

- [ ] `pages` が描画責務に専念し、Wails API や複雑な状態管理を直接持っていないか
- [ ] Wails API が `hooks/features/*` 経由に保たれているか
- [ ] `pages` から `wailsjs` / `store` を直接 import していないか
- [ ] 境界値を `unknown` + runtime parse で扱えているか
- [ ] adapter の責務が UI 層へ漏れていないか
- [ ] `lint:file -> 修正 -> 再実行 -> lint:frontend` の流れを満たしているか

## 4. Change-Specific Risks

- [ ] この change で特有の退行リスク 1: `%w` 必須判定が意図的な error 変換ケースまで誤検知しないか
- [ ] この change で特有の退行リスク 2: cleanup error 無視の例外規則が本流の握りつぶしを見逃さないか
- [ ] データ互換性 / 移行 / 再実行時のリスク: 新 rule 導入時に既存コードへの一括修正負荷が過大にならないか

## 5. Verification Notes

- 重点的に見るべきファイル: `tools/backendquality/**`, 追加する analyzer / rule 実装, 代表 fixture
- 重点的に見るべきテスト: package 境界 `return err`, `fmt.Errorf` `%w` 欠落, cleanup 例外ケース
- verify 時の補足メモ: wrap 必須境界の定義が spec と実装で一致しているか確認する

## 6. Adoption Notes

- `go test ./tools/backendquality/...` は成功した。
- `npm run backend:check` は成功し、既存テストや goimports 差分には影響しなかった。
- `npm run backend:lint` は `errorwrapcheck` により 106 件の既存違反を検出して失敗した。
- 既存違反は主に `pkg/infrastructure/llm`, `pkg/infrastructure/queue`, `pkg/config`, `pkg/task`, `pkg/summary`, `pkg/controller`, `cmd/parser` に集中している。
- 第一段の修正優先度は、公開境界 `return err` が多い package からとし、`do not ignore errors` は best-effort か本流かを切り分けながら段階的に潰す。
- `strings.Builder` / `bytes.Buffer` の `Write*` と `fmt.Print*` は best-effort 例外として analyzer 側で除外済みであり、残件は実コード側の wrap 不足または追加例外検討候補として扱う。
