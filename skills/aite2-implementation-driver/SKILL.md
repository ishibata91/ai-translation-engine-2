---
name: aite2-implementation-driver
description: AI Translation Engine 2 専用。設計済み change を実装し、backend/frontend の lint、typecheck、E2E、品質ゲートまで進める。設計文書を読みながら安全に実装するときに使う。
---

# AITE2 Implementation Driver

この skill は実装用。
目的は、設計差分をコードへ反映し、PJ 既定の品質ゲートを漏れなく通すこと。

## 使う場面
- `changes/<id>/` の設計が揃っている
- backend または frontend を実装する
- lint / typecheck / test / E2E まで完了させたい

## 入力
- `changes/<id>/tasks.md`
- 必要なら `ui.md` `scenarios.md` `logic.md`
- 関連コード

## 出力
- 実装差分
- 実行した品質確認の結果
- 未解決事項

## 実装順序
1. change 文書を読む
2. 影響ファイルを特定する
3. 小さく実装する
4. 変更ファイル単位で lint を潰す
5. 横断ゲートを実行する
6. 文書との差分や未完了を整理する

## 品質ゲート
- backend 変更: `backend:lint:file -> 修正 -> 再実行 -> lint:backend`
- frontend 変更: `lint:file -> 修正 -> 再実行 -> typecheck -> lint:frontend -> Playwright`

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- 大きい変更を一括で入れず、小さな変更単位で実装と確認を繰り返す
- 各ステップで完了条件、未解決事項、次の 1 手を明確にする
- 設計が曖昧なまま実装で補完しない
- 既存規約より局所最適を優先しない
- UI とロジックの責務を混ぜない
- 未検証の箇所は明示する
