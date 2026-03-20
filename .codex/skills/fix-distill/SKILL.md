---
name: fix-distill
description: AI Translation Engine 2 専用。再現条件、関連コード、既知観測、関連仕様を bugfix packet に蒸留する。bugfix flow の入口で事実整理をしたいときに使う。
---

# Fix Distill

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は bugfix 調査に必要な bugfix packet を作る skill。
再現条件と既知観測を整理し、次の調査方針を決めるための材料を返す。

## 制約
- 恒久修正を提案しない。
- 原因推定は行わない。
- 事実、再現条件、関連 artifact を分けて返す。

## やること
1. 対象の症状、再現条件、change を確認する。
2. 関連する docs、コード、context board、既知ログを読む。
3. 再現条件、観測済み事実、関連箇所、未観測箇所を整理する。
4. `fix-direction` が次工程へ渡す bugfix packet を返す。

## 原則
- bugfix packet を返した後の次工程は決めない
- review や docs handoff の判断は行わない

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
