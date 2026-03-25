---
name: impl-distill
description: AI Translation Engine 2 専用。確定済み artifacts を実装 packet に蒸留する。`ui.md` `scenarios.md` `logic.md` から `impl-direction` が次判断できる implementation packet を作りたいときに使う。
---

# Impl Distill

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は実装開始に必要な implementation packet を作る skill。
plan 側で確定済みの artifact を読み、実装開始に必要な packet を返す。
基本的にMCP経由で走査すること｡
## 許可される運用範囲
- 返却内容は implementation packet の蒸留に限り、仕様再解釈や設計判断は含めない。
- section 分割と `tasks.md` 生成の判断は `impl-direction` に委ねる。
- docs 同期判断は下流の review / sync 系判断材料に委ねる。
- 実装判断が必要な unknowns はそのまま返す。
- 出力正本は `changes/<id>/context_board/impl-distill.packet.json` とし、会話本文だけを正本にしない。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、`impl-distill.packet.validation.json` を出力する。
- validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する。

## やること
1. 対象 change と task を確認する。
2. `ui.md` `scenarios.md` `logic.md` と関連 docs / コードを必要最小限だけ読む。
3. 実装に必要な interface、entry point、module 候補、edit boundary、validation_commands を抽出する。
4. shared contract 候補と implementation を止める unknowns を分けて整理する。
5. `impl-direction` が次判断できる implementation packet を返す。

## packet 契約
- `invoked_skill`: `impl-distill`
- `invoked_by`: `impl-direction`
- `change`: 対象 change
- `task`: 実装対象
- `scope`: 走査した artifact とコード範囲
- `must_read`: worker が必ず読む path
- `interfaces`: 実装で守るべき contract
- `entry_points`: 着手点となる path / symbol
- `module_candidates`: モジュール/契約単位の section 候補
- `shared_contract_candidates`: 先に固定すべき shared contract 候補
- `edit_boundary`: 変更してよい境界
- `validation_commands`: worker が実行すべき検証コマンド。validation field はこれだけを使う
- `constraints`: 実装時の固定条件
- `acceptance`: 完了条件
- `unknowns`: 実装判断を止める論点
- `handoff`: 返却先 direction と must-read

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
- packet / validation schema の検証は `.codex/skills/scripts/validate-packet-contracts.ps1` を使う。
