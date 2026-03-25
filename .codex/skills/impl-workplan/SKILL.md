---
name: impl-workplan
description: AI Translation Engine 2 専用。`impl-distill` の implementation packet を読み込み、モジュール/契約単位の section plan と `changes/<id>/tasks.md` を生成する。frontend / backend work に渡す実装計画を固めたいときに使う。
---

# Impl Workplan

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-workplan` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は implementation packet を section 単位の実装計画へ変換する skill。
`impl-distill` が返した packet を読み、モジュール/契約単位の section plan、condensed brief、`changes/<id>/tasks.md` を作る。
基本的にMCP経由で走査すること｡

`Section Plan` `Work Order` `tasks.md format` の schema 正本は `references/templates.md` のみとし、
worker に渡す前に `owner / depends_on / shared_contract / condensed_brief / owned_paths / forbidden_paths / required_reading / validation_commands / acceptance`
を必ず固定する。
ここで固定する `shared_contract` は、worker が `owned_paths` の中だけで section を完了するために必要十分な契約でなければならない。

## 入力契約
- `impl-direction` から渡された implementation packet
- implementation packet の validation field は `validation_commands` のみ
- 対象 change
- `ui.md` `scenarios.md` `logic.md`
- 既存 `tasks.md`（resume / reroute で再計画する場合）
- 関連コード

## 許可される運用範囲
- 返却内容は section plan と `tasks.md` 生成に限り、設計変更やコード実装は含めない。
- section plan は 1 section = 1 owner の構成で扱う。
- frontend と backend の品質ゲートは section ごとに分離して扱う。
- 出力正本は `changes/<id>/context_board/impl-workplan.packet.json` とし、会話本文だけを正本にしない。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、`impl-workplan.packet.validation.json` を出力する。
- validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する。
- unresolved な shared contract が残る場合は、unknown を含む packet として返す。
- 完成扱いにできる section plan は、owner 確定、shared contract 固定、required field 充足を満たすものに限る。
- `impl-workplan` は `tasks.md` の section 契約と初期 status snapshot を定義する。runtime 中の status / 実装 / 検証更新は `impl-direction` が行ってよい。
- docs 同期判断は review / sync 系判断材料に委ねる。
- worker が `owned_paths` 内で compile / test / wiring を完了できる section を作る。
- constructor / DI / test stub / store contract の追従が別 path に必要なら、その依存を先行 section として分離するか、同一 owner の section へ吸収する。
- runtime wiring section は、呼び出される constructor signature と shared contract が先行 section で確定済みの場合だけ生成する。
- store / slice 契約を拡張する section は、その契約を利用する downstream section が blocked にならない粒度まで read/write 契約を固定する。

## 分割候補
- section は owner だけでなく layer boundary でも分割する
- backend では artifact / slice / workflow / controller を同一 section に混在させない
- frontend では page / hook / adapter / component を同一 section に混在させない
- shared contract の追加変更がある場合、それ自体を先行 section として独立させる
- 2 つ以上の package または feature directory をまたぐ section は broad とみなし再分割する
- backend と frontend の validation_commands を同一 section に含めてはならない
- constructor / DI / test stub の更新が別 path に必要な場合は、compile 境界で section を追加分割する
- downstream worker が `shared_contract が不足しているため owned_paths 内で完結できない` と判断しそうな場合は、section 生成をやめて unresolved に倒す

## やること
1. 対象 change と implementation packet を確認する。
2. `ui.md` `scenarios.md` `logic.md` と関連コードを必要最小限だけ読む。
3. モジュール/契約単位で section 候補を洗い出す。
4. 各 section について `section_id / title / owner / status / goal / depends_on / shared_contract / condensed_brief / owned_paths / forbidden_paths / required_reading / validation_commands / acceptance` を確定する。
5. 各 section ごとに、worker が `owned_paths` だけで compile / test / wiring を完了できるかを確認する。別 path の constructor / DI / test stub / store contract 追従が必要なら section を再分割する。
6. dispatch 順、shared contract 一覧、progress snapshot を整理する。
7. `references/templates.md` の `tasks.md format` に従い、`impl-workplan` 自身が `changes/<id>/tasks.md` を生成または更新する。
8. `impl-direction` がそのまま dispatch できる workplan packet を返す。

## packet 契約
- `change`: 対象 change
- `tasks_path`: 生成または更新した `changes/<id>/tasks.md`
- `progress_snapshot`: section ごとの初期または再計画後 status 一覧
- `shared_contracts`: section 着手前に固定済みの shared contract 一覧
- `dispatch_order`: section の実行順
- `sections`: `references/templates.md` の `Section Plan` schema に一致した一覧
- `work_orders`: `references/templates.md` の `Work Order` schema に一致した一覧
- `unresolved`: worker 起動を止める論点
- `handoff`: 返却先 direction と must-read

## 参照
- section plan と work order の雛形は `references/templates.md` を使う。
- 返答テンプレートは `references/response-template.md` を使う。

## 許可される動作
- `tasks.md` は impl lane の正本 artifact として扱う
- `tasks.md` の section 契約は `impl-workplan` が定義し、runtime 中の status と検証注記だけを `impl-direction` が更新する
- 再計画は同一 scope 内の section 再編として扱う
- section はファイル単位ではなくモジュール/契約単位で切る
- worker へ渡すのは owner と shared contract が確定した section に限る
- shared contract は workplan 側で固定して worker へ渡す
- condensed brief は `required_reading` の代替ではなく、worker が履歴全文を再読せず着手するための圧縮本文として固定する
- shared contract は API 名の列挙ではなく、worker が `owned_paths` 内だけで section を完了するための実装前提として固定する
- `depends_on` は参照順と compile 順が整合する形で定義する
- constructor / DI / test stub / contract 依存は dispatch 前に workplan 側で解消する
- implementation packet にある前提の範囲で workplan を組み立てる
- section plan を完成扱いにするのは unresolved が解消済みの場合に限る
