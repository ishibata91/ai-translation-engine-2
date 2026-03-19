---
name: aite2-review-guard
description: AI Translation Engine 2 専用。レビュー入口を整理する。設計レビューか実装レビューか迷うときに起動する。
---

# AITE2 Review Guard

この skill は review 系 skill の入口を整理し、最初に使う review skill と次の 1 手を決めるための司令塔 skill。

## 使う場面
- 設計レビューか実装レビューかを切り分けたい
- 旧名 `aite2-review-guard` でレビュー依頼が来た
- change / docs の整合確認か、実装差分レビューかを判断したい

## 手順
1. 対象が設計文書中心か、実装差分中心かを判定する。
2. 設計レビューなら `aite2-design-review-guard` を使う。
3. 実装レビューなら `aite2-implementation-review-guard` を使う。
4. 実装レビューで finding が出る場合は、呼び出し元 implementation skill が修正し、再度 `aite2-implementation-review-guard` を呼ぶ。

## 標準遷移
- 設計系 skill からの後続: `aite2-design-review-guard`
- backend / frontend / bug-fix 実装後の後続: `aite2-implementation-review-guard`

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 設計レビューと実装レビューを混ぜない
- 実装修正ループは review skill 自身ではなく呼び出し元 implementation skill が担う
