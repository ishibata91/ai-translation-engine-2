---
name: fix-analysis
description: AI Translation Engine 2 専用。bugfix flow の専用ログや観測出力を事実へ圧縮する optional skill。再現後のログから何が起きたかだけを返したいときに使う。
---

# Fix Analysis

この skill は専用ログや観測出力を構造化 JSON に圧縮し、事実だけを整理して返す skill。

## 使う場面
- 再現後のログを原因調査へ渡す前に整理したい
- ノイズの多い出力から起きた事実だけを抽出したい

## 入力契約
- 専用ログ
- 再現時の観測出力
- context board

## 手順
1. ログと観測出力を読む。
2. 時系列と重要イベントを抽出する。
3. 構造化 JSON と短い要約へ落とす。
4. 原因推定を混ぜず、何が起きたかだけを返す。

## 原則
- 原因推定はしない
- 事実、時系列、関連箇所だけを返す
- 次の skill が読める形式へ整える
- 次工程や review の起動は決めない

## 参照資料
- 出力形式は `references/templates.md` を使う。
