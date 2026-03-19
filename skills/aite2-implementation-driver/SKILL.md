---
name: aite2-implementation-driver
description: AI Translation Engine 2 専用。実装入口を振り分ける。「この change を実装して」と言われたときに frontend / backend のどちらの実装 skill を使うか決める。
---

# AITE2 Implementation Driver

この skill は frontend 実装か backend 実装かを切り分け、新しい実装 skill へ送るための移行用 skill。

## 使う場面
- 旧名 `aite2-implementation-driver` で実装依頼が来た
- frontend 実装か backend 実装かを切り分けたい
- 新しい実装 skill へ誘導したい
- 実装後のレビュー修正ループまで含めた入口を決めたい

## 手順
1. 実装対象が frontend か backend かを判定する。
2. frontend なら `aite2-frontend-implementation` を使う。
3. backend なら `aite2-backend-implementation` を使う。
4. 混在する場合は先に task を分割して担当 skill を決める。
5. 各 implementation skill が `aite2-implementation-review-guard` を後続で使う前提を共有する。

## 実装後の標準フロー
- 実装 skill は変更を入れた後に `aite2-implementation-review-guard` を呼ぶ
- finding があれば同じ implementation skill が修正する
- 修正後に再度 `aite2-implementation-review-guard` を呼ぶ
- 重大 / 中程度 finding がなくなるまで続ける

## 参照資料
- 実装の振り分け例は `references/examples.md` を読む。
- 新しい実装 skill の違いは `references/quality-checklist.md` を読む。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- frontend と backend の品質ゲートを 1 skill に混ぜない
- 各ステップで次に使う skill を明確にする
- 設計レビューと実装レビューを混同しない
