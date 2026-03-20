---
name: aite2-implementation-driver
description: AI Translation Engine 2 専用。実装入口を振り分ける。「この change を実装して」と言われたときに frontend / backend のどちらの実装 skill を使うか決める。
---

# AITE2 Implementation Driver

この skill は impl 系作業の指揮者として動き、task の切り分け、`aite2-explorer` による context 圧縮、実装 skill の起動、review skill の起動、finding の差し戻しまでを管理する orchestration skill。

## 使う場面
- 旧名 `aite2-implementation-driver` で実装依頼が来た
- frontend 実装か backend 実装かを切り分けたい
- `aite2-explorer` で実装前 context を圧縮したい
- 実装 skill を起動したい
- 実装完了後に review skill を起動したい
- 実装後のレビュー修正ループまで含めた入口を決めたい

## 手順
1. 実装依頼、`tasks.md`、関連する change 文書を読み、実装 task を最小単位へ分割する。
2. `aite2-explorer` に、関連する `changes/` `docs/` 既存コードの context packet を集めさせる。
3. 各 task が frontend / backend / mixed のどれかを判定する。
4. 単一領域なら該当 implementation skill を 1 つ `spawn_agent` する。
5. mixed なら frontend task と backend task に分割し、それぞれの implementation skill を別 `spawn_agent` する。
6. 実装完了後、統合差分を対象に `aite2-implementation-review-guard` を `spawn_agent` する。
7. review skill から `critical` または `medium` finding が返ったら、該当 implementation skill へ `send_input` で差し戻す。
8. `critical` と `medium` finding が 0 件になるまで review loop を続ける。

## 標準 orchestration フロー
1. 指揮者は orchestration 以外を行わない。自分で探索、実装、レビュー、検証をせず、task 分解、context 収集、担当判定、agent 間の受け渡し、終了判定だけを行う。
2. 最初の context 収集は `aite2-explorer` で必ず 1 回行う。
3. 実装は `aite2-frontend-implementation` または `aite2-backend-implementation` を使う。
4. review は `aite2-implementation-review-guard` を使う。
5. mixed task のときだけ implementation skill を並列起動する。
6. review skill は task ごとではなく、実装が出そろった時点の統合差分を対象に 1 回起動する。
7. docs 同期が必要な場合だけ `aite2-sync-docs` を使う。

## subagent 起動規約
- implementation 指揮者は、原則として下位 skill を subagent として起動する
- 下位 skill の起動時は、各 skill 配下の `agents/openai.yaml` に定義された subagent 設定を使う
- ローカル直実行は、ユーザーが delegation を明示的に止めた場合か、subagent 側設定が欠けていて継続不能な場合に限る
- context 圧縮は `aite2-explorer` を使う
- 実装は `aite2-frontend-implementation` または `aite2-backend-implementation` を使う
- review は `aite2-implementation-review-guard` を使う
- docs 同期が必要な場合は `aite2-sync-docs` を使う
- `aite2-explorer` には今回の対象 task、読むべき change、読むべき docs、読むべきコードを明示して渡す
- 実装 skill には所有ファイル範囲と必読 spec を明示して渡す
- review skill には change 文書、`docs/` 正本、差分、検証結果、前回 findings をまとめて渡す

## review loop の扱い
- review skill が `critical` または `medium` finding を返したら、指揮者が実装担当 skill を特定して差し戻す
- 実装 skill への差し戻しでは、前回 finding の解消確認を最優先 task として渡す
- 同一 finding が 2 周続けて残る場合はユーザーへエスカレーションする
- finding が `low` のみなら、未解決リスクとして整理して loop を終了してよい

## 参照資料
- 実装の振り分け例は `references/examples.md` を読む。
- 新しい実装 skill の違いは `references/quality-checklist.md` を読む。
- subagent への依頼文は `references/templates.md` を使う。

## 原則
- 作業は対話内でタスク化し、常に 1 ステップずつ進める
- frontend と backend の品質ゲートを 1 skill に混ぜない
- 各ステップで次に使う skill と担当 agent を明確にする
- 設計レビューと実装レビューを混同しない
- `aite2-explorer` を通さずに大きい change をそのまま実装 skill に渡さない
- mixed task でも review skill は統合差分に対して 1 回だけ立てる
- review skill の指摘を指揮者が要約せず、そのまま実装 skill に返す
