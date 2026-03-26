---
name: impl-backend-work
description: AI Translation Engine 2 専用。backend 所有範囲の実装だけを行う。work order に従って workflow / slice / gateway / runtime task を実装したいときに使う。
---

# Impl Backend Work

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-backend-work` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は backend 所有範囲の実装結果を返す skill。
section work order で指定された所有範囲だけを変更し、backend 品質ゲートを返す。
1 つの section を完了するか、blocked として差し戻し理由を返した時点で停止する。

## 使う場面
- `impl-direction` が `impl-workplan` の section plan に従って backend section work order を渡した
- `pkg/` や backend 層の owned paths だけを変更したい
- shared contract を守りつつ workflow / slice / gateway 実装を進めたい

## 入力契約
- `impl-direction` から渡された work order
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
- `docs/governance/backend-quality-gates/spec.md`

## 手順
1. work order の `section_id` `goal` `required_reading` `shared_contract` `condensed_brief` `progress_snapshot` を読む。
   - `design_principles` を確認し、SRP/SoC などの責務境界を壊さない実装方針を先に固定する。
2. 以下の規約およびアーキテクチャ定義に従って実装する。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. `docs/governance/backend-quality-gates/spec.md` を読む。
4. 自分の `owned_paths` だけで実装対象を確定する。`depends_on` は前提参照であり、依存 section の実装着手権限ではない。
5. 外部ライブラリ、フレームワーク、SDK、ミドルウェアの現行仕様に依存する実装や設定変更を行う場合は、`Context7 MCP` を使って対象ドキュメントを確認してから実装する。ライブラリ ID 解決と docs 取得を行い、推測で API や設定を決めない。
6. `owned_paths` 外の編集、別 section の責務、未固定 contract 解消が必要だと分かった時点で実装を止め、`references/templates.md` の `Section Result` と blocked 形式で `completed_scope` `remaining_gap` `noise_classification` `reroute_hint` を返す。
7. 小さな変更単位で実装する。
8. `backend:lint:file -> 修正 -> 再実行` を回す。
9. `lint:backend` を実行する。品質ゲートを通すために `owned_paths` 外の修正が必要になった場合は修正せず blocked を返す。失敗は `scope_failure` `external_validation_noise` `known_pre_existing_issue` に分類する。
10. structured diff 形式で `section_id` `result` `completed_scope` `remaining_gap` `changed_paths` `validation_result` `noise_classification` `reroute_hint` `unverified` `principle_adherence` を返す。
11. 返却後は停止し、次 section の開始判断を `impl-direction` に委ねる。

## レビュー修正ループ
- `impl-direction` から required delta が返ったら、その解消を最優先で実施する
- 前回 feedback に明示的に触れながら修正内容を返す
- review の起動判断は `impl-direction` に委ねる

## 参照資料
- 実装メモの雛形は `references/templates.md` を使う。
- 実装の進め方の例は `references/examples.md` を読む。
- 品質ゲート順序は `references/quality-checklist.md` を使う。
- `owned_paths` 外が必要になったときの返却形式も `references/templates.md` を使う。

## 許可される動作
- 自分の所有範囲に責任を持つ
- 実装範囲は 1 section の goal に限る
- `tasks.md` や他 work order は参照用にとどめ、自 section の実装に集中する
- `depends_on` は読み順の制約として扱い、編集権限は `owned_paths` に限る
- `owned_paths` 外の編集が必要な場合は blocked として返す
- blocked 返却では、どこまで完了したかと何が未解消かを必ず分ける
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 設計が曖昧な場合は実装で補完せず `remaining_gap` として返す
- 実装判断は既存規約と責務境界に合わせる
- `design_principles` に記載された SRP/SoC/DIP/OCP のうち該当分を満たすよう実装し、逸脱した場合は `remaining_gap` に理由を残す
- 外部依存の API や設定値は `Context7 MCP` で確認できる限り確認してから使う
- 他 agent の変更を巻き戻さない
- 未検証の箇所は明示する
- validation failure を外部ノイズと section failure で混ぜない
- section 完了後の次 section 判断は `impl-direction` に委ねる
