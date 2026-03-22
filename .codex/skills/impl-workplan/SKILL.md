---
name: impl-workplan
description: AI Translation Engine 2 専用。`impl-distill` の implementation packet を読み込み、モジュール/契約単位の section plan と `changes/<id>/tasks.md` を生成する。frontend / backend work に渡す実装計画を固めたいときに使う。
---

# Impl Workplan

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-workplan` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は implementation packet を section 単位の実装計画へ変換する skill。
`impl-distill` が返した packet を読み、モジュール/契約単位の section plan と `changes/<id>/tasks.md` を作る。
基本的にMCP経由で走査すること｡

`Section Plan` `Work Order` `tasks.md format` の schema 正本は `references/templates.md` のみとし、
worker に渡す前に `owner / depends_on / shared_contract / owned_paths / forbidden_paths / required_reading / validation_commands / acceptance`
を必ず固定する。

## 入力契約
- `impl-direction` から渡された implementation packet
- implementation packet の validation field は `validation_commands` のみ
- 対象 change
- `ui.md` `scenarios.md` `logic.md`
- 関連コード

## 制約
- 設計変更やコード実装は行わない。
- 1 section = 1 owner を崩さない。
- frontend と backend の品質ゲートを同一 section に混在させない。
- unresolved な shared contract が残る場合は worker 起動可能な計画を返さず、unknown として止める。
- owner 未確定、shared contract 未固定、required field 欠落の section plan は完成扱いにしない。
- `changes/<id>/tasks.md` の生成または更新は `impl-workplan` だけが行い、worker へ委譲しない。
- docs 同期判断は行わない。

## 分割候補
- section は owner だけでなく layer boundary でも分割する
- backend では artifact / slice / workflow / controller を同一 section に混在させない
- frontend では page / hook / adapter / component を同一 section に混在させない
- shared contract の追加変更がある場合、それ自体を先行 section として独立させる
- 2 つ以上の package または feature directory をまたぐ section は broad とみなし再分割する
- backend と frontend の validation_commands を同一 section に含めてはならない

## やること
1. 対象 change と implementation packet を確認する。
2. `ui.md` `scenarios.md` `logic.md` と関連コードを必要最小限だけ読む。
3. モジュール/契約単位で section 候補を洗い出す。
4. 各 section について `section_id / title / owner / goal / depends_on / shared_contract / owned_paths / forbidden_paths / required_reading / validation_commands / acceptance` を確定する。
5. dispatch 順と shared contract 一覧を整理する。
6. `references/templates.md` の `tasks.md format` に従い、`impl-workplan` 自身が `changes/<id>/tasks.md` を生成または更新する。
7. `impl-direction` がそのまま dispatch できる workplan packet を返す。

## packet 契約
- `change`: 対象 change
- `tasks_path`: 生成または更新した `changes/<id>/tasks.md`
- `shared_contracts`: section 着手前に固定済みの shared contract 一覧
- `dispatch_order`: section の実行順
- `sections`: `references/templates.md` の `Section Plan` schema に一致した一覧
- `work_orders`: `references/templates.md` の `Work Order` schema に一致した一覧
- `unresolved`: worker 起動を止める論点
- `handoff`: 次に使う skill と must-read

## 参照
- section plan と work order の雛形は `references/templates.md` を使う。
- 返答テンプレートは `references/response-template.md` を使う。

## 原則
- `tasks.md` は impl lane の正本 artifact として扱う
- `tasks.md` は `impl-workplan` だけが生成または更新する
- section はファイル単位ではなくモジュール/契約単位で切る
- section に placeholder owner や未固定 shared contract を残したまま worker へ流さない
- shared contract を worker に判断させない
- implementation packet に無い前提を勝手に足さない
- unresolved があるなら section plan を完成扱いしない
