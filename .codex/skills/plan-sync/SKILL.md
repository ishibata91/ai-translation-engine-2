---
name: plan-sync
description: AI Translation Engine 2 専用。changes 配下の設計差分から docs 正本へ昇格すべき内容を抽出し、安全に同期する。review 済みの仕様差分を canonical docs へ反映したいときに使う。
---

# Plan Sync

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `plan-sync` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は仕様差分のマージ用。
`changes/<id>/` を丸ごと写すのではなく、docs 正本へ昇格すべき仕様だけを抽出して統合する。

## 入力契約
- `plan-direction` `impl-direction` `fix-direction` などから渡された docs sync handoff
- 対象 change
- 反映先 docs
- change 側から docs へ昇格させる候補

## 使う場面
- `changes/<id>/` の設計差分を `docs/` へ反映したい
- 実装後に docs の正本を更新したい
- change 文書と docs の乖離を解消したい
- `logic.md` にしかない仕様断片を docs へ昇格させたい

## 主な入力
- `changes/<id>/ui.md`
- `changes/<id>/scenarios.md`
- `changes/<id>/logic.md` を仕様抽出元として必要に応じて参照する
- `changes/<id>/review.md` があれば同期時の注意点として参照する

## 全文同期しないもの
- `changes/<id>/logic.md`
- 実装詳細
- クラス構成や局所的ワークアラウンド
- 一時的な設計メモ
- 責務分担の比較検討

## docs へ昇格させる対象
- 対象集合
- 恒久的な振る舞いルール
- 正常系 / 異常系の成立条件
- 例外条件
- 永続化される事実
- 外部公開契約として残る前提

## docs へ昇格させない対象
- どこを正本にするかの設計理由
- 実装都合の分解方針
- 局所 workaround
- 暫定メモ
- 責務境界そのものの検討過程

## やること
1. `ui.md` と `scenarios.md` を確認する。
2. 反映先の `docs/` 正本を特定する。
3. 仕様差分が足りない場合のみ `logic.md` から docs 昇格候補を抽出する。
4. `review.md` がある場合は、未解決 finding を持ち込まないよう確認する。
5. 差分を統合し、冗長や矛盾を整理する。
6. docs を更新した後、change 側との整合を確認する。
7. `logic.md` から何を昇格させたかを短く要約する。

## 許可される動作
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- docs を正本として扱う
- change 文書は正本向けに統合して反映する
- `logic.md` はふるまい仕様の抽出元として扱う
- docs へ反映する内容は恒久仕様に絞る
- 仕様差分の意味が曖昧な場合は open question として返す
- 設計と実装の lane 判断は direction skill へ委ねる
- docs 正本への昇格に集中し、implementation 判断材料へ徹する

## 参照資料
- docs 昇格メモは `references/templates.md` を使う。
