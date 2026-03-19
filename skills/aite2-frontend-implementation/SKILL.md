---
name: aite2-frontend-implementation
description: AI Translation Engine 2 専用。frontend change を実装する。「画面を実装して」「frontend task を進めて」と言われたときに起動する。
---

# AITE2 Frontend Implementation

この skill は設計済み frontend change をコードへ反映し、frontend 品質ゲートまで通すための実装 skill。

## 使う場面
- `ui.md` や `scenarios.md` に沿って frontend を実装したい
- `frontend/src` 配下の task を進めたい
- typecheck、lint、Playwright まで完了させたい
- 実装後に `implementation-review-guard` と修正ループを回したい

## 必読 spec
- `docs/frontend/frontend-architecture/spec.md`
- `docs/frontend/frontend-coding-standards/spec.md`
- 補助: `docs/governance/playwright-quality-gate/spec.md`

## 手順
1. `tasks.md` と関連する `ui.md` `scenarios.md` を読む。
2. `docs/frontend/frontend-architecture/spec.md` と `docs/frontend/frontend-coding-standards/spec.md` を読む。
3. 実装対象と影響ファイルを特定する。
4. 小さな変更単位で実装する。
5. `lint:file -> 修正 -> 再実行` を回す。
6. `typecheck -> lint:frontend -> Playwright` を実行する。
7. `aite2-implementation-review-guard` で実装レビューを行う。
8. finding があれば同じ skill で修正し、再度 `aite2-implementation-review-guard` を呼ぶ。
9. 未完了、仕様差分、確認結果を整理する。

## レビュー修正ループ
- 実装後の自己確認で止めず、`aite2-implementation-review-guard` を必ず後続に置く
- `review-guard` から重大 / 中程度 finding が返ったら修正する
- 修正後に再レビューを行い、指摘ゼロまで繰り返す
- 同一論点が 2 周以上解消しない場合はユーザーへ確認する

## 参照資料
- 実装メモの雛形は `references/templates.md` を使う。
- 実装の進め方の例は `references/examples.md` を読む。
- 品質ゲート順序は `references/quality-checklist.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 各ステップで完了条件、未解決事項、次の 1 手を明確にする
- 設計が曖昧なまま実装で補完しない
- UI とロジックの責務を混ぜない
- 未検証の箇所は明示する
