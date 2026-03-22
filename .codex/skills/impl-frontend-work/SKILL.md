---
name: impl-frontend-work
description: AI Translation Engine 2 専用。frontend 所有範囲の実装だけを行う。work order に従って UI / frontend task を実装したいときに使う。
---

# Impl Frontend Work

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-frontend-work` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は frontend 所有範囲の実装結果を返す skill。
section work order で指定された所有範囲だけを変更し、frontend 品質ゲートを返す。
1 つの section を完了するか、blocked として差し戻し理由を返した時点で停止する。

## 使う場面
- `impl-direction` が `impl-workplan` の section plan に従って frontend section work order を渡した
- `frontend/src` 配下の owned paths だけを変更したい
- shared contract を守りつつ UI / frontend 実装を進めたい

## 入力契約
- `impl-workplan` から渡された work order
- `section_id`
- `goal`
- `depends_on`
- `progress_snapshot`
- shared contract
- condensed brief
- 所有ファイル範囲と禁止範囲
- 実行すべき品質ゲート
- この work order は単一 section だけを対象にする

## 必読 spec
- `docs/governance/architecture/spec.md`
- `docs/governance/backend-coding-standards/spec.md`
- `docs/frontend/frontend-coding-standards/spec.md`
- `docs/frontend/frontend-architecture/spec.md`

## 手順
1. work order の `section_id` `goal` `required_reading` `shared_contract` `condensed_brief` `progress_snapshot` を読む。
2. 以下の規約およびアーキテクチャ定義に従って実装する。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. `docs/frontend/frontend-architecture/spec.md` を読む。
4. 自分の `owned_paths` だけで実装対象を確定する。`depends_on` は前提参照であり、依存 section の実装着手権限ではない。
5. `owned_paths` 外の編集、別 section の責務、未固定 contract 解消が必要だと分かった時点で実装を止め、`references/templates.md` の `Section Result` と blocked 形式で `completed_scope` `remaining_gap` `noise_classification` `reroute_hint` を返す。
6. 小さな変更単位で実装する。
7. `lint:file -> 修正 -> 再実行` を回す。
8. `typecheck -> lint:frontend` を実行する。品質ゲートを通すために `owned_paths` 外の修正が必要になった場合は修正せず blocked を返す。失敗は `scope_failure` `external_validation_noise` `known_pre_existing_issue` に分類する。
9. structured diff 形式で `section_id` `result` `completed_scope` `remaining_gap` `changed_paths` `validation_result` `noise_classification` `reroute_hint` `unverified` を返す。
10. 返却後は停止し、次 section の開始判断を `impl-direction` に委ねる。

## レビュー修正ループ
- `impl-review` から required delta が返ったら、その解消を最優先で実施する
- 前回 feedback に明示的に触れながら修正内容を返す
- 自身では review を起動しない

## 参照資料
- 実装メモの雛形は `references/templates.md` を使う。
- 実装の進め方の例は `references/examples.md` を読む。
- 品質ゲート順序は `references/quality-checklist.md` を使う。
- `owned_paths` 外が必要になったときの返却形式も `references/templates.md` を使う。

## 原則
- 自分の所有範囲に責任を持つ
- 1 section の goal を越えて別 section の責務へ広げない
- `tasks.md` や他 work order を読んでも、自 section 以外の実装を開始しない
- `depends_on` は読み順の制約であり、他 section の編集許可ではない
- `owned_paths` 外の編集が必要なら自分で広げず blocked として返す
- blocked 返却では、どこまで完了したかと何が未解消かを必ず分ける
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 設計が曖昧なまま実装で補完しない
- UI とロジックの責務を混ぜない
- 他 agent の変更を巻き戻さない
- 未検証の箇所は明示する
- validation failure を外部ノイズと section failure で混ぜない
- section 完了後は停止し、次 section を自分で始めない
