## 1. Canonical Structure

- [x] 1.1 `openspec/specs` の正規 zone ディレクトリを作成し、`governance / frontend / controller / workflow / slice / runtime / artifact / gateway / foundation` の最終構造を用意する
- [x] 1.2 `architecture`、`spec-structure`、品質ゲート、テスト標準、ログ標準、全体要求などの共通文書を `governance` 配下へ移動する
- [x] 1.3 `artifact` と `foundation` の独立区分に対応する spec を正規パスへ揃える

## 2. Capability Reclassification

- [x] 2.1 `master-persona-ui`、`dashboard` など UI 責務の spec を `frontend` 配下へ移動する
- [x] 2.2 `queue`、`task`、`config`、`datastore`、`llm` などの共通 capability を `runtime` または `gateway` の正規パスへ移動する
- [x] 2.3 旧配置と新配置の二重管理を解消し、空ディレクトリや不要な旧パスを整理する

## 3. Mixed Spec Split

- [x] 3.1 `translation-flow-data-load` を frontend と workflow の 2 capability へ分割し、元 spec の混在要件を除去する
- [x] 3.2 `spec-structure` と `architecture` の本文を新しい canonical path と責務区分へ更新する
- [x] 3.3 補助文書と相互参照を新パスへ張り替える

## 4. Entry Points And Verification

- [x] 4.1 `AGENTS.md` の spec 参照先を新しい構造へ更新する
- [x] 4.2 `openspec/specs` の directory tree と参照切れを確認し、残件を解消する
- [x] 4.3 `tasks.md` と `review.md` を更新し、change を完了状態まで整理する
