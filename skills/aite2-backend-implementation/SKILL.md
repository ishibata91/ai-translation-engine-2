---
name: aite2-backend-implementation
description: AI Translation Engine 2 専用。backend change を実装する。「workflow を実装して」「backend task を進めて」と言われたときに起動する。
---

# AITE2 Backend Implementation

この skill は設計済み backend change をコードへ反映し、backend 品質ゲートまで通すための実装 skill。

## 使う場面
- `logic.md` に沿って backend を実装したい
- workflow / slice / runtime / artifact / gateway / controller の task を進めたい
- backend lint と品質ゲートまで完了させたい

## 必読 spec
- `docs/governance/backend-coding-standards/guide.md`
- `docs/governance/backend-quality-gates/spec.md`
- 補助: `docs/governance/architecture/spec.md`

## 手順
1. `tasks.md` と関連する `logic.md` を読む。
2. `docs/governance/backend-coding-standards/guide.md` と `docs/governance/backend-quality-gates/spec.md` を読む。
3. 必要なら `docs/governance/architecture/spec.md` で責務境界を再確認する。
4. 実装対象と影響ファイルを特定する。
5. 小さな変更単位で実装する。
6. `backend:lint:file -> 修正 -> 再実行` を回す。
7. `lint:backend` を実行する。
8. 未完了、仕様差分、確認結果を整理する。

## 参照資料
- 実装メモの雛形は `references/templates.md` を使う。
- 実装の進め方の例は `references/examples.md` を読む。
- 品質ゲート順序は `references/quality-checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 各ステップで完了条件、未解決事項、次の 1 手を明確にする
- 設計が曖昧なまま実装で補完しない
- 既存規約より局所最適を優先しない
- 責務境界違反を持ち込まない
- 未検証の箇所は明示する
