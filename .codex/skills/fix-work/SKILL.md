---
name: fix-work
description: AI Translation Engine 2 専用。bugfix flow で確定した修正方針だけを実装する。`sub_implementer` が fix scope と修正方針に従って最小修正を入れるときに使う。
---

# Fix Work

この skill は bugfix flow の実装修正役として、確定した修正方針だけを実装する。

## 使う場面
- `fix-trace` と `fix-direction` が fix scope を確定済み
- 最小修正だけを安全に反映したい
- 原因調査を再開せず、指定作業だけを進めたい

## 入力契約
- bugfix 指揮者から渡された fix plan
- context board
- 所有ファイル範囲
- 実行すべき品質ゲート

## 手順
1. fix plan と context board を読む。
2. 指示された修正だけを実装する。
3. 必要な品質ゲートを実行する。
4. 実装内容と未解消事項を bugfix 指揮者へ返す。

## 原則
- 新しい原因調査を始めない
- 指示されていない refactor を混ぜない
- 自分の所有範囲だけを変更する

## 参照資料
- 修正返却には `references/templates.md` を使う。
