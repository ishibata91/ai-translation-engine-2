---
name: aite2-sync-docs
description: AI Translation Engine 2 専用。changes 配下の UI / scenario 差分文書を docs 配下の正本へ反映し、仕様差分を安全にマージするための簡易 skill。設計が固まった後や実装完了後に、change 文書から docs を同期したいときに使う。
---

# AITE2 Sync Docs

この skill は仕様差分のマージ用。
今は簡易草案として、何を扱うかだけを定義する。

## 使う場面
- `changes/<id>/` の設計差分を `docs/` へ反映したい
- 実装後に docs の正本を更新したい
- change 文書と docs の乖離を解消したい

## 同期対象
- `changes/<id>/ui.md`
- `changes/<id>/scenarios.md`

## 同期対象外
- `changes/<id>/logic.md`
- 実装詳細
- クラス構成や局所的ワークアラウンド
- 一時的な設計メモ

## やること
- `changes/<id>/ui.md` と `changes/<id>/scenarios.md` を確認する
- 反映先の `docs/` 正本を特定する
- 差分を統合し、冗長や矛盾を整理する
- docs を更新した後、change 側との整合を確認する

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- docs を正本として扱う
- change 文書をそのまま丸写しせず、正本向けに統合する
- UI / scenario だけを同期し、logic は同期しない
- 実装都合の詳細を docs へ持ち込まない
- 仕様差分の意味が曖昧なら勝手にマージしない
