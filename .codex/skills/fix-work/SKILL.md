---
name: fix-work
description: AI Translation Engine 2 専用。bugfix flow で確定した修正方針だけを実装する。fix scope と修正方針に従って最小修正を反映したいときに使う。
---

# Fix Work

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-work` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は確定した修正方針だけを実装する skill。

## 使う場面
- `fix-direction` が fix scope を確定済み
- 最小修正だけを安全に反映したい
- 原因調査を再開せず、指定作業だけを進めたい

## 入力契約
- bugfix 指揮者から渡された fix plan

- 所有ファイル範囲
- 確定済み fix scope
- state summary
- 実行すべき品質ゲート

## 手順
1. 渡された fix plan（パケット）を読む。
2. 以下の規約およびアーキテクチャ定義に従って実装する。
   - `docs/governance/architecture/spec.md`
   - `docs/governance/backend-coding-standards/spec.md`
   - `docs/frontend/frontend-coding-standards/spec.md`
3. 外部ライブラリ、フレームワーク、SDK、ミドルウェアの API や設定値を修正する場合は、`Context7 MCP` を使って対象ドキュメントを確認してから実装する。ライブラリ ID 解決と docs 取得を行い、記憶や推測だけで修正しない。
4. 指示された修正だけを実装する。
5. 必要な品質ゲートを実行し、失敗は `scope_failure` `external_validation_noise` `known_pre_existing_issue` に分類する。
6. 実装内容、`completed_scope`、検証結果、`remaining_gap`、未解消事項を bugfix 指揮者へ返す。

## 許可される動作
- 実装は確定済み fix scope に集中し、新しい原因調査は `fix-direction` へ戻す
- 指示された修正を中心に実装し、未指示の refactor は分離して扱う
- 自分の所有範囲だけを変更する
- 外部依存の API や設定値は `Context7 MCP` で確認できる限り確認してから修正する
- DB メンテナンス、データ補正、再投入、再生成で解消できる問題に対して、起動時 self-heal や無理やり帳尻を合わせるコードを追加しない
- 未解消事項は `remaining_gap` として明示し、外部ノイズと混ぜない
- review と再投入の判断は `fix-direction` 経由で受ける

## 参照資料
- 修正返却には `references/templates.md` を使う。
