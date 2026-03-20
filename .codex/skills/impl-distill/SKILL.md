---
name: impl-distill
description: AI Translation Engine 2 専用。確定済み artifacts を実装 packet に蒸留する。`ui.md` `scenarios.md` `logic.md` `tasks.md` から interface-only packet を作りたいときに使う。
---

# Impl Distill

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `impl-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は実装開始に必要な implementation packet を作る skill。
plan 側で確定済みの artifact を読み、実装開始に必要な packet を返す。

## 制約
- 仕様の再解釈や設計判断を行わない。
- worker 分割は行わない。
- docs 同期判断は行わない。
- 実装判断が必要な unknowns はそのまま返す。

## やること
1. 対象 change と task を確認する。
2. `ui.md` `scenarios.md` `logic.md` `tasks.md` と関連 docs / コードを必要最小限だけ読む。
3. 実装に必要な interface、entry point、edit boundary、quality gate を抽出する。
4. shared contract 候補と implementation を止める unknowns を分けて整理する。
5. `impl-workplan` がそのまま使える implementation packet を返す。

## packet 契約
- `task`: 実装対象
- `scope`: 走査した artifact とコード範囲
- `must_read`: worker が必ず読む path
- `interfaces`: 実装で守るべき contract
- `entry_points`: 着手点となる path / symbol
- `edit_boundary`: 変更してよい境界
- `owned_paths_candidates`: worker 分割候補
- `quality_gates`: 実行すべき検証
- `constraints`: 実装時の固定条件
- `acceptance`: 完了条件
- `unknowns`: 実装判断を止める論点
- `handoff`: 次に使う skill と must-read

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
