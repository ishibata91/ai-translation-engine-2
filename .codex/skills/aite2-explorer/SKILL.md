---
name: aite2-explorer
description: AI Translation Engine 2 専用。関連仕様、関連コード、既知の再現条件、観測ログを圧縮し、次の skill に渡す context packet を作る。
---

# AITE2 Explorer

## 目的
- 指示された範囲だけを走査し、次の skill が読める最小の context packet に圧縮する。

## 制約
- 自分では実装しない。
- 事実整理と判断を混ぜない。
- board は要約面であり、生ログ置き場にしない。
- 次の skill が読む必要のある情報だけを残す。
- subagent として起動される前提で使い、`agents/openai.yaml` の profile 設定に従う。

## やること
1. 対象 task と change の有無を確認する。
2. 指示された `changes/` `docs/` `frontend/` `pkg/` の範囲だけ読む。
3. 関連 docs、関連コード、再現条件、観測ログを事実ベースで整理する。
4. `changes/<id>/context_board/` を初期化または更新する。
5. 次の skill が読むべき最小 context packet を返す。

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
