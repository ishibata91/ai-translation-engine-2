---
name: aite2-explorer
description: AI Translation Engine 2 専用。関連仕様、関連コード、既知の再現条件、観測ログを圧縮し、次の skill に渡す context packet を作る。
---

# AITE2 Explorer

この skill は各フローの入口で context を圧縮する探索役として動き、関連 docs、関連コード、既知の再現条件、観測ログを必要最小限の packet に整理する。

## 使う場面
- design / implementation / bugfix / ui-refine の入口で context を集めたい
- change 配下の `context_board/` を初期化したい
- 次の skill に渡す入力を圧縮したい
- 調査結果を事実ベースでまとめたい

## 入力契約
- 対象 task または change id
- 読むべき `changes/` `docs/` `frontend/` `pkg/` の範囲
- 既知の再現条件または観測ログ
- 次に起動する skill 名

## 手順
1. 対象 task と change の有無を確認する。
2. 関連する `changes/` `docs/` `frontend/` `pkg/` から必要な範囲だけ読む。
3. 期待挙動、実挙動、関連コード、未確定事項を切り分ける。
4. `changes/<id>/context_board/` を初期化または更新する。
5. 次の skill が読むべき最小 context packet を返す。

## 出力契約
- `current_context.md`
  - 対象 task
  - 関連 docs
  - 関連コード
  - 既知の事実
  - 制約
- `handoff.md`
  - 現在の担当
  - 次に起動する skill
  - 完了条件
  - 未確定事項
- 必要なら追加メモ
  - 再現条件
  - ログ要約
  - 仕様差分メモ

## 原則
- 自分では実装しない
- 原因判断と事実整理を混ぜない
- board は要約面であり、生ログ置き場にしない
- 次の skill が読む必要のある情報だけを残す
