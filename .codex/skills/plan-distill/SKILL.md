---
name: plan-distill
description: AI Translation Engine 2 専用。差分仕様作成と実装計画確定のために、要求、既存 docs、既存コード、制約を蒸留する。plan 系 skill の入口として planning packet を作りたいときに使う。
---

# Plan Distill

この skill は `ctx_loader` 用の設計蒸留 skill。
要求と既存 artifact を整理し、次の plan skill が設計判断を始められる planning packet を返す。

## 制約
- 自分では UI / シナリオ / ロジックを確定しない。
- 実装 packet は作らない。
- 事実、制約、未確定事項を混ぜない。

## やること
1. 要求と対象 change を確認する。
2. 関連する `changes/` `docs/` `frontend/` `pkg/` を必要最小限だけ読む。
3. 現在ある artifact、関連 spec、制約、未確定論点を整理する。
4. `plan-ui` `plan-scenario` `plan-logic` のどれに渡すべきかを handoff に書く。
5. planning packet を返す。

## packet 契約
- `request`: 今回の要求と対象 change
- `current_artifacts`: 既存の `ui.md` `scenarios.md` `logic.md` `tasks.md` の有無と状態
- `relevant_specs`: 先に読むべき docs とコード
- `constraints`: 守るべき仕様、構造、運用制約
- `open_decisions`: 次の plan skill で決める論点
- `conflicts`: 既存 artifact 間の矛盾
- `draft_targets`: 新規作成または更新すべき artifact
- `handoff`: 次に使う skill と必読パス

## 参照
- 返答テンプレートは `references/response-template.md` を使う。
