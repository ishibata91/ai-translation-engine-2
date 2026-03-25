---
name: plan-distill
description: AI Translation Engine 2 専用。差分仕様作成と実装計画確定のために、要求、既存 docs、既存コード、制約を整理し、planning packet を作る。次の plan skill が判断を始めるための材料をまとめたいときに使う。
---

# Plan Distill

> **起動確認**: このスキルが起動されたら、まず `invoked_skill` が `plan-distill` であることを確認する。不一致の場合は作業を開始せずエラーを返す。

この skill は設計判断に必要な planning packet を作る skill。
要求と既存 artifact を整理し、次の plan skill が設計判断を始められる材料を返す。
基本的にMCP経由で走査すること｡
## 許可される運用範囲
- 返却内容は planning packet の蒸留に限り、UI / シナリオ / ロジックの確定は下流 plan skill へ委ねる。
- 実装 packet 化の判断は対応する direction へ委ねる。
- 事実、制約、未確定事項を分けて返す。

## やること
1. 要求と対象 change を確認する。
2. 関連する `changes/` `docs/` `frontend/` `pkg/` を必要最小限だけ読む。
3. 現在ある artifact、関連 spec、制約、未確定論点を整理する。
4. `plan-direction` が次判断できるよう、未確定論点と必読パスを handoff に書く。
5. planning packet を返す。

## packet 契約
- `request`: 今回の要求と対象 change
- `current_artifacts`: 既存の `ui.md` `scenarios.md` `logic.md` `tasks.md` の有無と状態
- `relevant_specs`: 先に読むべき docs とコード
- `constraints`: 守るべき仕様、構造、運用制約
- `open_decisions`: 次の plan skill で決める論点
- `conflicts`: 既存 artifact 間の矛盾
- `draft_targets`: 新規作成または更新すべき artifact
- `handoff`: 返却先 direction と必読パス

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
