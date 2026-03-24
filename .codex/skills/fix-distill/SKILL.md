---
name: fix-distill
description: AI Translation Engine 2 専用。再現条件、関連コード、既知観測、関連仕様を bugfix packet に蒸留する。bugfix flow の入口で事実整理をしたいときに使う。
---

# Fix Distill

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は bugfix 調査に必要な bugfix packet を作る skill。
再現条件と既知観測を整理し、次の調査方針を決めるための condensed bugfix packet を返す。
基本的にMCP経由で走査すること｡

## 制約
- 恒久修正を提案しない。
- 原因推定は行わない。
- 事実、再現条件、関連 artifact、state summary seed を分けて返す。
- 出力正本は `changes/<id>/context_board/fix-distill.packet.json` とし、会話本文だけを正本にしない。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、`fix-distill.packet.validation.json` を出力する。
- validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する。

## やること
1. 対象の症状、再現条件、change を確認する。
2. 関連する docs、コード、既知ログを読む。
3. 再現条件、観測済み事実、関連箇所、未観測箇所、fix scope の未確定点を整理する。
4. `fix-direction` がそのまま `State Summary` へ反映できる condensed bugfix packet を返す。

## 原則
- bugfix packet を返した後の次工程は決めない
- review や docs handoff の判断は行わない
- `must_read` の列挙だけにせず、要約本文で次 skill が着手できる形にする

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
- packet / validation schema の検証は `.codex/skills/scripts/validate-packet-contracts.ps1` を使う。
