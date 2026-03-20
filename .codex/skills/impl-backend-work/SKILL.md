---
name: impl-backend-work
description: AI Translation Engine 2 専用。backend 所有範囲の実装だけを行う。work order に従って workflow / slice / gateway / runtime task を実装したいときに使う。
---

# Impl Backend Work

この skill は backend 所有範囲の実装結果を返す skill。
work order で指定された所有範囲だけを変更し、backend 品質ゲートを返す。

## 使う場面
- `impl-workplan` から backend work order を受け取った
- `pkg/` や backend 層の owned paths だけを変更したい
- shared contract を守りつつ workflow / slice / gateway 実装を進めたい

## 入力契約
- `impl-workplan` から渡された work order
- shared contract
- 所有ファイル範囲と禁止範囲
- 実行すべき品質ゲート

## 必読 spec
- `docs/governance/backend-coding-standards/guide.md`
- `docs/governance/backend-quality-gates/spec.md`
- 補助: `docs/governance/architecture/spec.md`

## 手順
1. work order の `required_reading` と `shared_contract` を読む。
2. `docs/governance/backend-coding-standards/guide.md` と `docs/governance/backend-quality-gates/spec.md` を読む。
3. 必要なら `docs/governance/architecture/spec.md` で責務境界を再確認する。
4. 自分の所有ファイル範囲だけで実装対象と影響ファイルを特定する。
5. 小さな変更単位で実装する。
6. `backend:lint:file -> 修正 -> 再実行` を回す。
7. `lint:backend` を実行する。
8. structured diff 形式で実装結果、実行コマンド、未検証箇所を返す。

## レビュー修正ループ
- `impl-review` から required delta が返ったら、その解消を最優先で実施する
- 前回 feedback に明示的に触れながら修正内容を返す
- 自身では review を起動しない

## 参照資料
- 実装メモの雛形は `references/templates.md` を使う。
- 実装の進め方の例は `references/examples.md` を読む。
- 品質ゲート順序は `references/quality-checklist.md` を使う。

## 原則
- 自分の所有範囲に責任を持つ
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 設計が曖昧なまま実装で補完しない
- 既存規約より局所最適を優先しない
- 責務境界違反を持ち込まない
- 他 agent の変更を巻き戻さない
- 未検証の箇所は明示する
