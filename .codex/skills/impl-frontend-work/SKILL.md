---
name: impl-frontend-work
description: AI Translation Engine 2 専用。frontend 所有範囲の実装だけを行う。`sub_implementer` が work order に従って UI / frontend task を実装するときに使う。
---

# Impl Frontend Work

この skill は `sub_implementer` 用の frontend 実装 skill。
work order で指定された所有範囲だけを変更し、frontend 品質ゲートを返す。

## 使う場面
- `impl-workplan` から frontend work order を受け取った
- `frontend/src` 配下の owned paths だけを変更したい
- shared contract を守りつつ UI / frontend 実装を進めたい

## 入力契約
- `impl-workplan` から渡された work order
- shared contract
- 所有ファイル範囲と禁止範囲
- 実行すべき品質ゲート

## 必読 spec
- `docs/frontend/frontend-architecture/spec.md`
- `docs/frontend/frontend-coding-standards/spec.md`

## 手順
1. work order の `required_reading` と `shared_contract` を読む。
2. `docs/frontend/frontend-architecture/spec.md` と `docs/frontend/frontend-coding-standards/spec.md` を読む。
3. 自分の所有ファイル範囲だけで実装対象と影響ファイルを特定する。
4. 小さな変更単位で実装する。
5. `lint:file -> 修正 -> 再実行` を回す。
6. `typecheck -> lint:frontend` を実行する。
7. structured diff 形式で実装結果、実行コマンド、未検証箇所を返す。

## レビュー修正ループ
- `impl-review` から required delta が返ったら、その解消を最優先で実施する
- 前回 feedback に明示的に触れながら修正内容を返す
- 自身では review を起動しない

## 参照資料
- 実装メモの雛形は `references/templates.md` を使う。
- 実装の進め方の例は `references/examples.md` を読む。
- 品質ゲート順序は `references/quality-checklist.md` を使う。

## 原則
- `sub_implementer` として自分の所有範囲に責任を持つ
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 設計が曖昧なまま実装で補完しない
- UI とロジックの責務を混ぜない
- 他 agent の変更を巻き戻さない
- 未検証の箇所は明示する
