---
name: aite2-log-analyze
description: AI Translation Engine 2 専用。bugfix flow で専用ログや観測出力を構造化 JSON に圧縮し、何が起きたかだけを返す。bugfix flow の log analysis が必要なときに起動する。
---

# AITE2 Log Analyze

この skill は bugfix flow のログ解析役として、専用ログや観測出力を構造化 JSON に圧縮し、事実だけを整理して返す。

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

## 参照資料
- 出力形式は `references/templates.md` を使う。
