---
name: aite2-log-parser
description: AI Translation Engine 2 専用。bugfix flow で debugger 専用ログや観測出力を構造化 JSON に圧縮し、何が起きたかだけを返す。bugfix flow の log parsing が必要なときに起動する。
---

# AITE2 Log Parser

この skill は bugfix flow の log parser 役として、debugger 専用ログや観測出力を構造化 JSON に圧縮し、事実だけを整理して返す。

## 使う場面
- 再現後のログを debugger へ渡す前に整理したい
- ノイズの多い出力から起きた事実だけを抽出したい

## 入力契約
- debugger 専用ログ
- 再現時の観測出力
- context board

## subagent 実行前提
- この skill は bugfix 指揮者から subagent として起動される前提で使う
- subagent 起動時は `agents/openai.yaml` の profile 設定を使う

## 手順
1. ログと観測出力を読む。
2. 時系列と重要イベントを抽出する。
3. 構造化 JSON と短い要約へ落とす。
4. 原因推定を混ぜず、何が起きたかだけを返す。

## 原則
- 原因推定はしない
- 事実、時系列、関連箇所だけを返す
- debugger が次に読める形式へ整える

## 参照資料
- 出力形式は `references/templates.md` を使う。
