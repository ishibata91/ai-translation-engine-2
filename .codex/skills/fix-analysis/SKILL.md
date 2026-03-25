---
name: fix-analysis
description: AI Translation Engine 2 専用。bugfix flow の専用ログや観測出力を事実へ圧縮する optional skill。再現後のログから何が起きたかだけを返したいときに使う。
---

# Fix Analysis

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-analysis` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は専用ログや観測出力を構造化 JSON に圧縮し、事実だけを整理して返す skill。
ログの場所はlogs/今日の日付.jsonlとする

## 使う場面
- 再現後のログを原因調査へ渡す前に整理したい
- ノイズの多い出力から起きた事実だけを抽出したい

## 入力契約
- 専用ログ
- 再現時の観測出力
- 直前の state summary


## 手順
1. ログと観測出力を読む。
2. 時系列と重要イベントを抽出する。
3. 構造化 JSON、短い要約、summary 更新用の差分へ落とす。
4. 原因推定を混ぜず、何が起きたかだけを返す。

## 許可される動作
- 原因推定を含めず、事実、時系列、関連箇所を返す
- `fix-direction` が次判断できる形式へ整える
- full log 再掲ではなく、state summary に反映する差分を中心に返す
- 次工程や review の判断材料に徹する

## 参照資料
- 出力形式は `references/templates.md` を使う。
