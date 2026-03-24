---
name: fix-trace
description: AI Translation Engine 2 専用。bugfix 調査で原因仮説と検証用 tracing 計画を作る。観測出力の追加で原因候補を絞り込みたいときに使う。
---

# Fix Trace

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `fix-trace` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は原因仮説、調査計画、調査用ログ配置計画、再現後の原因絞り込みを返す skill。
絶対に恒久修正を行わないこと｡他スキルのサブエージェントを呼び出したりしないこと｡ログの実際の追加は `fix-logging` が行う｡
出力正本は `changes/<id>/context_board/fix-trace.packet.json` とし、packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行して `fix-trace.packet.validation.json` を出力する。
validator fail 時は 1 回だけ自己再試行し、それでも fail なら invalid packet と validation artifact を残して終了する。

## 使う場面
- bugfix flow で最初の原因仮説を立てたい
- 再現前にログを配置したい
- 再現後のログと観測事実から原因候補を絞り込みたい

## 入力契約
- `fix-direction` または `fix-distill` から渡された bugfix packet
- `fix-direction` が保持する state summary
- 関連する docs
- 関連コード
- 再現条件
- 構造化された観測事実

## 手順
1. 渡されたbugfix packetから既知の症状と再現条件を読む。
2. 原因仮説と優先調査箇所を整理する。
3. 再現後は構造化された観測事実を読み、原因候補を狭める。
4. 観測ログが必要かを判断し、必要な場合は「仕込む対象ファイルと観測ポイント一覧」をパケットに含めて、`fix-direction` へ `fix-logging` の起動を要求して返す。ログの実装はこのスキルでは行わない。
5. `fix-direction` が state summary を更新できるよう、`current_hypothesis` `unknowns` `current_scope` `next_action` を明示して返す。
6. `fix-direction` へ fix plan の前提となる原因整理を返す。

## 原則
- 恒久修正は行わない
- ログ実装は行わない（`fix-logging` に委譲する）
- 原因仮説と観測事実を分けて返す
- full history の要約ではなく、state summary 更新に必要な差分だけを返す
- 次工程や review の起動は決めず、fix plan の前提事実だけを返す
- 必要観測未充足の間は `fix-work` を起動してよいとは返さない

## 参照資料
- handoff には `references/templates.md` を使う。
