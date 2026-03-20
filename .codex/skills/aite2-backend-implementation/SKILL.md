---
name: aite2-backend-implementation
description: AI Translation Engine 2 専用。backend change を実装する。「workflow を実装して」「backend task を進めて」と言われたときに起動する。
---

# AITE2 Backend Implementation

この skill は backend を担当する `coder` 向けの実装手順を定義し、設計済み backend change をコードへ反映して backend 品質ゲートまで通すための skill。

## 使う場面
- `logic.md` に沿って backend を実装したい
- workflow / slice / runtime / artifact / gateway / controller の task を進めたい
- orchestration された `coder` に backend task を渡したい
- backend lint と品質ゲートまで完了させたい
- 実装後に `implementation-review-guard` と修正ループを回したい

## coder 入力契約
- 対象 change と対象 task
- 所有ファイル範囲
- 必読 spec と補助 spec
- 実行すべき品質ゲート
- 完了時に返す報告形式

## 必読 spec
- `docs/governance/backend-coding-standards/guide.md`
- `docs/governance/backend-quality-gates/spec.md`
- 補助: `docs/governance/architecture/spec.md`

## 手順
1. 指揮者から渡された `tasks.md` と関連する `logic.md` を読む。
2. `docs/governance/backend-coding-standards/guide.md` と `docs/governance/backend-quality-gates/spec.md` を読む。
3. 必要なら `docs/governance/architecture/spec.md` で責務境界を再確認する。
4. 自分の所有ファイル範囲だけで実装対象と影響ファイルを特定する。
5. 小さな変更単位で実装する。
6. `backend:lint:file -> 修正 -> 再実行` を回す。
7. `lint:backend` を実行する。
8. 指揮者へ実装結果、実行コマンド、未検証箇所を返す。

## レビュー修正ループ
- `reviewer` から `critical` / `medium` finding が返ったら、その解消を最優先で実施する
- 前回 finding に明示的に触れながら修正内容を返す
- 修正後に `reviewer` の再レビューを受け、`critical` / `medium` が 0 件になるまで繰り返す
- 同一論点が 2 周以上解消しない場合は指揮者へエスカレーションする
- `coder` 自身は `reviewer` を起動しない。review 起動は指揮者が行う

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
- `coder` role の `worker` として自分の所有範囲に責任を持つ
- 他 agent の変更を巻き戻さない
- 未検証の箇所は明示する
